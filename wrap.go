package terrors

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"text/tabwriter"

	"github.com/go-faster/errors"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		pkg, _ := GetPackageAndFuncFromFuncName(runtime.FuncForPC(pc).Name())
		return FormatCaller(pkg, file, line)
	}
}

type wrapError struct {
	msg   string
	err   error
	frame Frame
	event *zerolog.Event
}

// var _ Framer = (*wrapError)(nil)

func (e *wrapError) Root() error {
	return e.err
}

func (e *wrapError) Frame() Frame {
	return e.frame
}

func (e *wrapError) Info() []any {
	return []any{e.msg}
}

func (e *wrapError) Event(ctx context.Context, gv func(*zerolog.Event) *zerolog.Event) error {
	e.event = gv(e.event)
	return e
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, v rune) { errors.FormatError(e, s, v) }

const (
	zerolog_info_key = "____info____"
)

func (e *wrapError) FormatError(p errors.Printer) (next error) {
	p.Print("ERROR[" + e.msg + "]")

	defer func() {
		e.event = nil
	}()

	if p.Detail() && (e.event != nil) {
		pkg, fn, file, line := e.frame.Location()
		l := zerolog.New(&printWriter{p})
		ev := e.event.Err(e.err).Str("caller", FormatCaller(pkg, file, line)).Str("function", fn).Stack()
		l.Err(nil).Dict(zerolog_info_key, ev).Send()
	}

	return e.err
}

func (e *wrapError) Unwrap() error {
	return e.err
}

// Wrap error with message and caller.
func Wrap(err error, message string) *wrapError {
	return WrapWithCaller(err, message, 1)
}

// Wrapf wraps error with formatted message and caller.
func Wrapf(err error, format string, a ...interface{}) *wrapError {
	return WrapWithCaller(err, fmt.Sprintf(format, a...), 1)
}

func WrapWithCaller(err error, message string, frm int) *wrapError {
	return &wrapError{msg: message, err: err, frame: Caller(frm + 1), event: zerolog.Dict()}
}

type printWriter struct {
	errors.Printer
}

func (p *printWriter) Write(b []byte) (int, error) {
	dat := map[string]interface{}{}

	err := json.Unmarshal(b, &dat)
	if err != nil {
		return 0, err
	}

	if _, ok := dat[zerolog_info_key]; ok {
		buf := bytes.NewBuffer(nil)

		wrt := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

		wrtfunc := func(t string, k string, v interface{}) error {
			if v == nil {
				return nil
			}
			if k == "error" {
				k = "parent"
			}
			_, err := wrt.Write([]byte(fmt.Sprintf("%s%s\t= %+v\n", t, k, v)))
			return err
		}

		if info, ok := dat[zerolog_info_key].(map[string]interface{}); ok {
			for k, v := range info {
				if err = wrtfunc("", k, v); err != nil {
					return 0, err
				}
			}
		}

		err := wrt.Flush()
		if err != nil {
			return 0, err
		}
		p.Printf("\n%s\n", buf.String())
	}

	return len(b), nil
}
