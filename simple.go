package terrors

import (
	"github.com/go-faster/errors"
)

// New returns an error that formats as the given text.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func New(text string) error {
	return WrapWithCaller(nil, text, errors.Caller(1))
}
