package terrors

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"context"
	"fmt"
	"runtime"
	"strings"

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
	pkg, _, filestr, linestr := e.Frame().Location()

	if e.err == nil {
		return fmt.Sprintf("ERROR[%s, %s, %s:%d]", e.msg, pkg, filestr, linestr)
	}

	errd := e.err.Error()

	arrow := "⏩"

	if !strings.Contains(errd, arrow) {
		arrow += "❌"
	}

	return fmt.Sprintf("ERROR[%s, %s, %s:%d] %s %s", e.msg, pkg, filestr, linestr, arrow, errd)
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

func (c *wrapError) MarshalZerologObject(e *zerolog.Event) (err error) {
	for _, ev := range c.event {
		*e = *ev(e)
	}
	return nil
}
