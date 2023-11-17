package validator_test

import (
	"testing"

	"github.com/ardikabs/go-stdlib/pkg/errs"
	"github.com/ardikabs/go-stdlib/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {

	v := validator.New()
	v.Check(true, errs.Parameter("field"), errs.Code("value_equal"), "field must not be empty")
	assert.Nil(t, v.Valid())

	v.Check(false, errs.Parameter("field"), errs.Code("value_non_equal"), "field must not be empty")
	assert.NotNil(t, v.Valid())
}

func TestValidatorPanic(t *testing.T) {
	v := validator.New()

	assert.Panics(t, func() {
		v.Check(false, errs.Parameter("first_name"))
	})

	assert.Panics(t, func() {
		v.AddError(errs.Parameter("last_name"))
	})

	assert.Panics(t, func() {
		v.AddError(errs.Parameter("last_name"), nil)
	})
}
