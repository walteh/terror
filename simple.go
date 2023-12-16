package terrors

import (
	"fmt"
)

// New returns an error that formats as the given text.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func New(text string) *wrapError {
	return WrapWithCaller(nil, text, 1)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func Errorf(format string, a ...any) *wrapError {
	return WrapWithCaller(nil, fmt.Sprintf(format, a...), 1)
}
