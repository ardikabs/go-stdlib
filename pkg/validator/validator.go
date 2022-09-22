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

// Check checking the given validation condition, followed with parameter and arguments which must contains one of the followings
// errs.Code, string, or error
func (v *Validator) Check(ok bool, param errs.Parameter, args ...interface{}) {
	if !argsCheck(args...) {
		panic("validator.Check: must be contains one of the followings `errs.Code`, `string`, or `error`")
	}

	if !ok {
		v.AddError(param, args...)
	}
}

// AddError add error for the validation with given parameter and arguments which must contains one of the followings
// errs.Code, string, or error
func (v *Validator) AddError(param errs.Parameter, args ...interface{}) {
	if !argsCheck(args...) {
		panic("validator.AddError: must be contains one of the followings `errs.Code`, `string`, or `error`")
	}

	if _, exist := v.stash[param]; !exist {
		v.stash[param] = true
		v.err.Append(param, errs.E(args...))
	}
}

func argsCheck(args ...interface{}) bool {
	if len(args) == 0 {
		return false
	}

	for _, arg := range args {
		switch arg.(type) {
		case errs.Code, string, error:
			continue
		default:
			return false
		}
	}

	return true
}
