package terrors

import "errors"

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
//
// Into is type-safe alternative to As.
func Into[T error](err error) (val T, ok bool) {
	ok = errors.As(err, &val)
	return val, ok
}

func IsRecoverable(err error) (bool, *Recovery) {
	// look for any recoverable error in the chain
	chain := GetChain(err)
	for _, e := range chain {
		if werr, ok := e.(*wrapError); ok {
			if werr.recovery != nil {
				return true, werr.recovery
			}
		}
	}

	return false, nil
}
