package terrors

import "errors"

type Framer interface {
	error
	Root() error
	Frame() Frame
	// Event() *zerolog.Event
	// Info() []any
}

// Cause returns first recorded Frame.
// func Cause(err error) (oerr error, f Frame, i []any, r bool) {
// 	for {
// 		we, ok := err.(Framer)
// 		if !ok {
// 			return err, f, i, r
// 		}
// 		f = we.Frame()
// 		// i = we.Info()
// 		r = r || ok
// 		oerr = err

// 		err = we.Root()
// 		if err == nil {
// 			return oerr, f, i, r
// 		}
// 	}
// }

func Cause2(err error) (f Framer, r bool) {
	for {
		we, ok := err.(Framer)
		if !ok {
			return
		}

		r = r || ok
		f = we

		err = we.Root()
		if err == nil {
			return
		}
	}
}

func ListCause(err error) ([]Framer, bool) {
	var frames []Framer

	for {
		we, ok := err.(Framer)
		if !ok {
			return frames, ok
		}

		frames = append(frames, we)

		err = we.Root()
		if err == nil {
			return frames, ok
		}
	}
}

func FirstCause(err error) (Framer, bool) {
	for {
		if err == nil {
			return nil, false
		}
		frm, ok := err.(Framer)
		if !ok {
			err = errors.Unwrap(err)
		} else {
			return frm, true
		}
	}
}
