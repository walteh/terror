package terrors

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"

	"github.com/go-faster/errors"
)

type wrapError struct {
	msg   string
	err   error
	frame Frame
}

var _ Framer = (*wrapError)(nil)

func (e *wrapError) Root() error {
	return e.err
}

func (e *wrapError) Frame() Frame {
	return e.frame
}

func (e *wrapError) Info() []any {
	return []any{e.msg}
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, v rune) { errors.FormatError(e, s, v) }

func (e *wrapError) FormatError(p errors.Printer) (next error) {
	p.Print(e.msg)
	e.frame.Format(p)
	return e.err
}

func (e *wrapError) Unwrap() error {
	return e.err
}

// Wrap error with message and caller.
func Wrap(err error, message string) error {
	return WrapWithCaller(err, message, Caller(1))
}

// Wrapf wraps error with formatted message and caller.
func Wrapf(err error, format string, a ...interface{}) error {
	return WrapWithCaller(err, fmt.Sprintf(format, a...), Caller(1))
}

func WrapWithCaller(err error, message string, frm Frame) error {
	return &wrapError{msg: message, err: err, frame: frm}
}

var _ Framer = &wrapError{}
