package terrors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/terrors"
)

func TestCause(t *testing.T) {
	err1 := fmt.Errorf("1")
	erra := terrors.Wrap(err1, "wrap 2")
	errb := terrors.Wrap(erra, "wrap3")

	_, v, ok := terrors.Cause(errb)
	if !ok {
		t.Error("unexpected false")
		return
	}

	assert.Equal(t, v, erra.(terrors.Framer).Frame())

}
