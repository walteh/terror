package terrors

import (
	"fmt"

	"github.com/go-faster/errors"
)

type TracableError interface {
	error
	Trace() error
}

var _ TracableError = (*allocError)(nil)

// errorString is a trivial implementation of error.
type allocError struct {
	s     string
	frame *errors.Frame
}

func New(str string) TracableError {
	return &allocError{s: str, frame: nil}
}

// New returns an error that formats as the given text.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func NewTraced(text string) error {
	if !errors.Trace() {
		return &allocError{text, nil}
	}
	ofs := errors.Caller(1)
	return &allocError{text, &ofs}
}

func NewDelay(text string) *allocError {
	return &allocError{text, nil}
}

func (e *allocError) SetFrame(callerOffset int) {
	if !errors.Trace() {
		return
	}
	ofs := errors.Caller(callerOffset)
	e.frame = &ofs
}

func (e *allocError) Error() string { return e.s }

func (e *allocError) Format(s fmt.State, v rune) { errors.FormatError(e, s, v) }

func (e *allocError) FormatError(p errors.Printer) (next error) {
	p.Print(e.s)
	if e.frame != nil {
		e.frame.Format(p)
	}
	return nil
}

func (e *allocError) Frame() errors.Frame {
	if e.frame == nil {
		return errors.Frame{}
	}
	return *e.frame
}

func (e *allocError) Root() error {
	return nil
}

func (e *allocError) Trace() error {
	e.SetFrame(2)
	return e
}
