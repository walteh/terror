package terrors

import (
	"errors"
	"slices"
)

// Into finds the first error in err's chain that matches target type T, and if so, returns it.
//
// Into is type-safe alternative to As.
func Into[T error](err error) (val T, ok bool) {
	ok = errors.As(err, &val)
	return val, ok
}

func IsRecoverable(err error) (bool, *Recovery) {
	chain := GetChain(err)

	// we want to get the deepest recoverable error in the chain
	slices.Reverse(chain)

	for _, e := range chain {
		if werr, ok := e.(*wrapError); ok {
			if werr.recovery != nil {
				return true, werr.recovery
			}
		}
	}

	return false, nil
}
