package terrors_test

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-faster/errors"
	"github.com/rs/zerolog"
	"github.com/walteh/terrors"
)

func TestIs(t *testing.T) {
	ctx := context.Background()

	ctx = zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger().WithContext(ctx)

	err1 := terrors.New("1")
	erra := terrors.Wrap(err1, "wrap 2").Event(func(e *zerolog.Event) *zerolog.Event {
		return e.Str("keyabc", "valueabc").Caller(1)
	})
	errb := terrors.Wrap(erra, "wrap3").Event(func(e *zerolog.Event) *zerolog.Event {
		return e.Str("key", "value").Caller(1)
	})
	erro := errors.Opaque(err1)
	errco := terrors.Wrap(erro, "opaque")
	err3 := terrors.New("3")

	poser := &poser{"either 1 or 3", func(err error) bool {
		return err == err1 || err == err3
	}}

	testCases := []struct {
		err    error
		target error
		match  bool
	}{
		{nil, nil, true},
		{nil, err1, false},
		{err1, nil, false},
		{err1, err1, true},
		{erra, err1, true},
		{errb, err1, true},
		{errco, erro, true},
		{errco, err1, false},
		{erro, erro, true},
		{err1, err3, false},
		{erra, err3, false},
		{errb, err3, false},
		{poser, err1, true},
		{poser, err3, true},
		{poser, erra, false},
		{poser, errb, false},
		{poser, erro, false},
		{poser, errco, false},
		{errorUncomparable{}, errorUncomparable{}, true},
		{errorUncomparable{}, &errorUncomparable{}, false},
		{&errorUncomparable{}, errorUncomparable{}, true},
		{&errorUncomparable{}, &errorUncomparable{}, false},
		{errorUncomparable{}, err1, false},
		{&errorUncomparable{}, err1, false},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			if got := errors.Is(tc.err, tc.target); got != tc.match {
				t.Errorf("Is(%v, %v) = %v, want %v", tc.err, tc.target, got, tc.match)
			} else {
				cause, ok := terrors.Cause2(tc.err)
				if !ok {
					// the test case was not designted to be printed

				} else {
					fmt.Println(terrors.FormatErrorCaller(cause, "cause", false))
				}
			}

		})
	}
}

type poser struct {
	msg string
	f   func(error) bool
}

func (p *poser) Error() string     { return p.msg }
func (p *poser) Is(err error) bool { return p.f(err) }
func (p *poser) As(err interface{}) bool {
	switch x := err.(type) {
	case **poser:
		*x = p
	case *errorT:
		*x = errorT{}
	case **os.PathError:
		*x = &os.PathError{}
	default:
		return false
	}
	return true
}

func TestAs(t *testing.T) {
	var errT errorT
	var errP *os.PathError
	var timeout interface{ Timeout() bool }
	var p *poser
	_, errF := os.Open("non-existing")

	testCases := []struct {
		err    error
		target interface{}
		match  bool
	}{{
		nil,
		&errP,
		false,
	}, {
		terrors.Wrap(errorT{}, "pittied the fool"),
		&errT,
		true,
	}, {
		errF,
		&errP,
		true,
	}, {
		errors.Opaque(errT),
		&errT,
		false,
	}, {
		errorT{},
		&errP,
		false,
	}, {
		errWrap{nil},
		&errT,
		false,
	}, {
		&poser{"error", nil},
		&errT,
		true,
	}, {
		&poser{"path", nil},
		&errP,
		true,
	}, {
		&poser{"oh no", nil},
		&p,
		true,
	}, {
		terrors.New("err"),
		&timeout,
		false,
	}, {
		errF,
		&timeout,
		true,
	}, {
		terrors.Wrap(errF, "path error"),
		&timeout,
		true,
	}}
	for i, tc := range testCases {
		name := fmt.Sprintf("%d:As(Errorf(..., %v), %v)", i, tc.err, tc.target)
		t.Run(name, func(t *testing.T) {
			match := errors.As(tc.err, tc.target)
			if match != tc.match {
				t.Fatalf("errors.As(%T, %T): got %v; want %v", tc.err, tc.target, match, tc.match)
			}
			if !match {
				return
			}
			if tc.target == nil {
				t.Fatalf("non-nil result after match")
			}
		})
	}
}

func TestAsValidation(t *testing.T) {
	var s string
	testCases := []interface{}{
		nil,
		(*int)(nil),
		"error",
		&s,
	}
	err := terrors.New("error")
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%T(%v)", tc, tc), func(t *testing.T) {
			defer func() {
				recover()
			}()
			if errors.As(err, tc) {
				t.Errorf("As(err, %T(%v)) = true, want false", tc, tc)
				return
			}
			t.Errorf("As(err, %T(%v)) did not panic", tc, tc)
		})
	}
}

func TestUnwrap(t *testing.T) {
	err1 := terrors.New("1")
	erra := terrors.Wrap(err1, "wrap 2")
	erro := errors.Opaque(err1)

	testCases := []struct {
		err  error
		want error
	}{
		{nil, nil},
		{errWrap{nil}, nil},
		{err1, nil},
		{erra, err1},
		{terrors.Wrap(erra, "wrap 3"), erra},

		{erro, nil},
		{terrors.Wrap(erro, "opaque"), erro},
	}
	for _, tc := range testCases {
		if got := errors.Unwrap(tc.err); got != tc.want {
			t.Errorf("Unwrap(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestOpaque(t *testing.T) {
	got := fmt.Sprintf("%v", terrors.Wrap(errors.Opaque(errorT{}), "foo"))
	want := "ERROR[msg=foo][pkg=walteh/terrors_test][file=wrap_test.go:233] 👉 ❌ errorT"
	if got != want {
		t.Errorf("error without Format: got %v; want %v", got, want)
	}

	got = fmt.Sprintf("%v", terrors.Wrap(errors.Opaque(errorD{}), "foo"))
	want = "ERROR[msg=foo][pkg=walteh/terrors_test][file=wrap_test.go:239] 👉 ❌ errorD"
	if got != want {
		t.Errorf("error with Format: got %q; want %q", got, want)
	}
}

type errorT struct{}

func (errorT) Error() string { return "errorT" }

type errorD struct{}

func (errorD) Error() string { return "errorD" }

func (errorD) FormatError(p errors.Printer) error {
	p.Print("errorD")
	p.Detail()
	p.Print("detail")
	return nil
}

type errWrap struct{ error }

func (errWrap) Error() string { return "wrapped" }

func (errWrap) Unwrap() error { return nil }

type errorUncomparable struct {
	_ interface{} // forbid compare
}

func (errorUncomparable) Error() string {
	return "uncomparable error"
}

func (errorUncomparable) Is(target error) bool {
	_, ok := target.(errorUncomparable)
	return ok
}
