package terrors

import "fmt"

// New returns an error that formats as the given text.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func New(text string) error {
	return WrapWithCaller(nil, text, Caller(1))
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
//
// The returned error contains a Frame set to the caller's location and
// implements Formatter to show this information when printed with details.
func Errorf(format string, a ...any) error {
	return WrapWithCaller(nil, fmt.Sprintf(format, a...), Caller(1))
}
