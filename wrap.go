package terrors

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sort"
	"text/tabwriter"

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
	event []func(*zerolog.Event) *zerolog.Event
}

func (e *wrapError) Root() error {
	return e.err
}

func (e *wrapError) Frame() Frame {
	return e.frame
}

func (e *wrapError) Info() []any {
	return []any{e.msg}
}

func (e *wrapError) Event(gv func(*zerolog.Event) *zerolog.Event) error {
	if gv != nil {
		e.event = append(e.event, gv)
	}
	return e
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, v rune) { FormatError(e, s, v) }

const (
	zerolog_info_key = "error_info"
)

func (e *wrapError) FormatError(p Printer) (next error) {
	p.Print("ERROR[" + e.msg + "]")

	if p.Detail() && (e.event != nil) {
		d := zerolog.Dict()
		for _, ev := range e.event {
			d = ev(d)
		}
		var w1 io.Writer
		w1 = &printWriter{p}
		if d.GetCtx() != nil {
			w1 = zerolog.MultiLevelWriter(w1, zerolog.Ctx(d.GetCtx()))
		}
		l := zerolog.New(w1)
		l.Err(nil).Dict(zerolog_info_key, d).Send()
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
	frme := Caller(frm + 1)

	return &wrapError{msg: message, err: err, frame: frme, event: []func(e *zerolog.Event) *zerolog.Event{
		func(e *zerolog.Event) *zerolog.Event {
			pkg, fn, file, line := frme.Location()
			e = e.Str("caller", FormatCaller(pkg, file, line)).Str("function", fn)
			if err != nil {
				e = e.Err(err)
			}
			return e
		},
	}}
}

type printWriter struct {
	Printer
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
				k = "chain"
			}
			_, err := wrt.Write([]byte(fmt.Sprintf("%s%s\t= %+v\n", t, k, v)))
			return err
		}

		if info, ok := dat[zerolog_info_key].(map[string]interface{}); ok {
			keys := []string{}
			for k := range info {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				if err = wrtfunc("", k, info[k]); err != nil {
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
