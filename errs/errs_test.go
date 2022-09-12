package errs_test

import (
	"fmt"
	"testing"

	"github.com/ardikabs/golib/errs"
	"github.com/stretchr/testify/assert"
)

func TestE(t *testing.T) {

	t.Run("new errs.Error", func(t *testing.T) {
		err := errs.E(errs.Other, errs.Code("another_code"), fmt.Errorf("some new error"))
		assert.NotNil(t, err)
		assert.Equal(t, errs.Other, errs.GetKind(err))
		assert.Equal(t, "some new error", err.Error())
	})

	t.Run("stacked error", func(t *testing.T) {
		err := errs.E(errs.Other, errs.Code("another_code"), errs.Parameter("param"))
		err = errs.E(errs.Validation, err)
		assert.Equal(t, errs.Validation, errs.GetKind(err))

		e, ok := err.(*errs.Error)
		assert.True(t, ok)
		assert.Equal(t, errs.Parameter("param"), e.Param)
		assert.Equal(t, errs.Code("another_code"), e.Code)
	})

	t.Run("string error", func(t *testing.T) {
		err := errs.E(errs.Internal, "internal server error")
		assert.Equal(t, "internal server error", err.Error())
	})
}
