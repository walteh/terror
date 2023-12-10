package terrors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
)

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

func FormatErrorCaller(err error, verbose bool) string {
	caller := ""
	var str string
	if frm, ok := Cause2(err); ok {
		pkg, _, filestr, linestr := frm.Frame().Location()
		caller = FormatCaller(pkg, filestr, linestr)
		caller = caller + " - "
		err = frm
	}

	if verbose {
		str = fmt.Sprintf("%+v\n", err)
		str = strings.Replace(str, "\n", "\n\n", 1)
	} else {
		str = fmt.Sprintf("%s", err)
	}

	prev := ""
	// replace any string that contains "*.Err" with a bold red version using regex
	str = regexp.MustCompile(`\S+\.Err\S*`).ReplaceAllStringFunc(str, func(s string) string {
		prev += color.New(color.FgRed, color.Bold).Sprint(s) + " -> "
		return ""
	})

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
