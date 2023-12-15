package terrors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/rs/zerolog"
)

func (e *wrapError) Detail() string {
	srtwrite := strings.Builder{}
	w1 := zerolog.New(&stringWriter{srtwrite})
	ed := w1.Err(e)
	for _, ev := range e.event {
		ed = ev(ed)
	}
	ed.Send()
	return srtwrite.String()
}

func FormatJsonForDetail(b []byte, ignored []string, priority []string) (string, error) {
	dat := map[string]any{}

	err := json.Unmarshal(b, &dat)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)

	wrt := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

	wrtfunc := func(t string, k string, v any) error {
		if v == nil {
			return nil
		}
		if v == "" || v == nil {
			return nil
		}
		_, err := wrt.Write([]byte(fmt.Sprintf("%s%s\t= %q\n", t, k, fmt.Sprintf("%v", v))))
		return err
	}

	priorityKeys := []string{}
	normal := []string{}
	for k := range dat {
		if slices.Contains(ignored, k) {
			continue
		}
		if slices.Contains(priority, k) {
			priorityKeys = append(priorityKeys, k)
		} else {
			normal = append(normal, k)
		}
	}

	slices.Sort(priorityKeys)
	slices.Sort(normal)

	for _, k := range priorityKeys {
		if err = wrtfunc("", k, dat[k]); err != nil {
			return "", err
		}
	}

	// if len(priorityKeys) > 0 {
	// 	_, err := wrt.Write([]byte("\n"))
	// 	if err != nil {
	// 		return "", err
	// 	}
	// }

	for _, k := range normal {
		if err = wrtfunc("", k, dat[k]); err != nil {
			return "", err
		}
	}

	err = wrt.Flush()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

type stringWriter struct {
	strings.Builder
}

func (s *stringWriter) Write(b []byte) (int, error) {
	str, err := FormatJsonForDetail(b, []string{"level"}, []string{"chain", "caller"})
	if err != nil {
		return 0, err
	}

	return s.Builder.WriteString(str)
}
