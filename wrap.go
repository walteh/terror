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
		w1 := zerolog.New(&printWriter{p})
		ed := w1.Err(nil)
		for _, ev := range e.event {
			ed = ev(ed)
		}
		// ctx := ed.GetCtx()
		// // if background context, then remove it
		// if ctx != nil && reflect.ValueOf(ctx).Type().String() != "context.todoCtx" {
		// 	w1 = w1.Output(zerolog.Ctx(ed.GetCtx()))
		// 	ed = w1.Err(nil)
		// 	for _, ev := range e.event {
		// 		ed = ev(ed)
		// 	}
		// }
		ed.Send()
	}

	return nil
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
			e = e.Str("caller", FormatCaller(pkg, file, line)).Str("function", fn).Ctx(context.TODO())
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

	keys := []string{}
	for k := range dat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {

		if k == "level" || dat[k] == nil || dat[k] == "" {
			continue
		}

		if err = wrtfunc("", k, dat[k]); err != nil {
			return 0, err
		}
	}

	err = wrt.Flush()
	if err != nil {
		return 0, err
	}
	p.Printf("\n%s\n", buf.String())

	return len(b), nil
}

type checker struct {
}

func (c *wrapError) MarshalZerologObject(e *zerolog.Event) (err error) {
	for _, ev := range c.event {
		ev(e)
	}
	return nil
}

// func init() {
// 	zerolog.ErrorMarshalFunc = func(err error) any {
// 		if err == nil {
// 			return nil
// 		}

// 		if e, ok := err.(*wrapError); ok {
// 			return e
// 		}

// 		return err
// 	}
// }
