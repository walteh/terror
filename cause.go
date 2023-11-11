package terrors

import "github.com/go-faster/errors"

type Framer interface {
	Root() error
	Frame() errors.Frame
}

// Cause returns first recorded Frame.
func Cause(err error) (oerr error, f errors.Frame, r bool) {
	for {
		we, ok := err.(Framer)
		if !ok {
			return err, f, r
		}
		f = we.Frame()
		r = r || ok
		oerr = err

		err = we.Root()
		if err == nil {
			return oerr, f, r
		}
	}
}
