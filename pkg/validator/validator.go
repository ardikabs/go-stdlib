package validator

import (
	"github.com/ardikabs/golib/pkg/errs"
)

type Validator struct {
	stash map[errs.Parameter]bool
	err   errs.ValidationErrors
}

func New() *Validator {
	return &Validator{
		stash: make(map[errs.Parameter]bool),
	}
}

func (v *Validator) Valid() error {
	if len(v.err) == 0 {
		return nil
	}

	return errs.E(errs.Validation, v.err)
}

func (v *Validator) Check(ok bool, param errs.Parameter, args ...interface{}) {
	if len(args) == 0 {
		panic("validator.Check: must be followed with arguments like `errs.Code`, `string`, or `error`")
	}

	if !ok {
		v.AddError(param, args...)
	}
}

func (v *Validator) AddError(param errs.Parameter, args ...interface{}) {
	if len(args) == 0 {
		panic("validator.AddError: must be followed with arguments like `errs.Code`, `string`, or `error`")
	}

	if _, exist := v.stash[param]; !exist {
		v.stash[param] = true
		v.err.Append(param, errs.E(args...))
	}
}
