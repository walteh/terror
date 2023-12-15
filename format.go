package terrors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
)

// A Formatter formats error messages.
type Formatter interface {
	error

	// FormatError prints the receiver's first error and returns the next error in
	// the error chain, if any.
	FormatError(p Printer) (next error)
}

// A Printer formats error messages.
//
// The most common implementation of Printer is the one provided by package fmt
// during Printf (as of Go 1.13). Localization packages such as golang.org/x/text/message
// typically provide their own implementations.
type Printer interface {
	// Print appends args to the message output.
	Print(args ...interface{})

	// Printf writes a formatted string.
	Printf(format string, args ...interface{})

	// Detail reports whether error detail is requested.
	// After the first call to Detail, all text written to the Printer
	// is formatted as additional detail, or ignored when
	// detail has not been requested.
	// If Detail returns false, the caller can avoid printing the detail at all.
	Detail() bool
}

func FileNameOfPath(path string) string {
	tot := strings.Split(path, "/")
	if len(tot) > 1 {
		return tot[len(tot)-1]
	}

	return path
}
func FormatCaller(pkg, path string, number int) string {
	return fmt.Sprintf("%s %s:%s", pkg, color.New(color.Bold).Sprint(FileNameOfPath(path)), color.New(color.FgHiRed, color.Bold).Sprintf("%d", number))
}

func ExtractErrorDetail(err error) string {
	if frm, ok := Cause2(err); ok {
		return frm.Detail()
	}

	return ""
}

func FormatErrorCaller(err error, verbose bool) string {
	caller := ""
	dets := ""
	var str string
	if frm, ok := Cause2(err); ok {
		pkg, _, filestr, linestr := frm.Frame().Location()
		caller = FormatCaller(pkg, filestr, linestr)
		caller = caller + " - "
		err = frm
		if verbose {
			dets = frm.Detail()
		}
	}

	if verbose && dets != "" {
		str = fmt.Sprintf("%s\n\n%s\n", err.Error(), dets)
	} else {
		str = fmt.Sprintf("%s", err.Error())
	}

	prev := ""
	// replace any string that contains "*.Err" with a bold red version using regex
	// str = regexp.MustCompile(`\S+\.Err\S*`).ReplaceAllStringFunc(str, func(s string) string {
	// 	prev += color.New(color.FgRed, color.Bold).Sprint(s) + " -> "
	// 	return ""
	// })

	return fmt.Sprintf("%s%s%s", caller, prev, color.New(color.FgRed).Sprint(str))
}

func FormatErrorCallerGoFaster(err error) string {
	caller := ""
	// the way go-faster/errors works is that you need to wrap to get the frame, so we do that here in case it has not been wrapped
	if frm, ok := errors.Cause(errors.Wrap(err, "tmp")); ok {
		_, filestr, linestr := frm.Location()
		caller = FormatCaller("", filestr, linestr)
		caller = caller + " - "
	}
	str := fmt.Sprintf("%+s", err)
	prev := ""
	// replace any string that contains "*.Err" with a bold red version using regex
	str = regexp.MustCompile(`\S+\.Err\S*`).ReplaceAllStringFunc(str, func(s string) string {
		prev += color.New(color.FgRed, color.Bold).Sprint(s) + " -> "
		return ""
	})

	return fmt.Sprintf("%s%s%s", caller, prev, color.New(color.FgRed).Sprint(str))
}
