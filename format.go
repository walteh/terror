package terrors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
)

func FormatCaller(pkg, path string, number int) string {
	tot := strings.Split(path, "/")
	if len(tot) > 1 {
		last := tot[len(tot)-1]
		// secondLast := tot[len(tot)-2]
		// thirdLast := tot[len(tot)-3]

		path = last

		// return fmt.Sprintf("%s/%s %s:%s", thirdLast, secondLast, color.New(color.Bold).Sprint(last), color.New(color.FgHiRed, color.Bold).Sprintf("%d", number))
	}

	return fmt.Sprintf("%s %s:%s", pkg, color.New(color.Bold).Sprint(path), color.New(color.FgHiRed, color.Bold).Sprintf("%d", number))
}

func FormatErrorCaller(err error) string {
	caller := ""
	var str string
	// the way go-faster/errors works is that you need to wrap to get the frame, so we do that here in case it has not been wrapped
	if frm, ok := Cause2(err); ok {
		pkg, _, filestr, linestr := frm.Frame().Location()
		caller = FormatCaller(pkg, filestr, linestr)
		caller = caller + " - "
		str = fmt.Sprintf("%+s", frm)
	} else {
		str = fmt.Sprintf("%+s", err)
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
