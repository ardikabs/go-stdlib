package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ardikabs/golib/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestErrs(t *testing.T) {

	t.Run("no arguments", func(t *testing.T) {
		assert.Panics(t, func() {
			_ = errs.E()
		})
	})

	t.Run("unexpected argument", func(t *testing.T) {
		var any interface{}
		err := errs.E(any)
		assert.NotNil(t, err)

		_, ok := err.(*errs.Error)
		assert.False(t, ok, "expected error should not be an errs.Error type, instead of a common error from fmt.Errorf")
	})

	t.Run("check error kind alias", func(t *testing.T) {
		assert.Equal(t, "other_error", errs.Other.String())
		assert.Equal(t, "I/O_error", errs.IO.String())
	})

	t.Run("Unwrap the error should return the unwrapped error", func(t *testing.T) {
		var errX = fmt.Errorf("error X")

		err := errs.E(
			errs.Code("internal"),
			errs.Internal,
			fmt.Errorf("internal error from X: %w", errX),
		)
		assert.True(t, errors.Is(errors.Unwrap(err), errX))
		assert.True(t, errs.KindIs(errs.Internal, err))
	})
}

func TestKindIs(t *testing.T) {

	testcases := []struct {
		err  error
		kind errs.Kind
		want bool
	}{
		{
			err:  nil,
			kind: errs.Invalid,
			want: false,
		},
		{
			err:  errs.E(errs.Internal, "new internal error"),
			kind: errs.Internal,
			want: true,
		},
		{
			err:  errs.E("some error"),
			kind: errs.Other,
			want: true,
		},
		{
			err:  errs.E("some error"),
			kind: errs.NotExist,
			want: false,
		},
		{
			err:  errs.E("nesting", errs.E(errs.Internal)),
			kind: errs.Internal,
			want: true,
		},
		{
			err:  errs.E("nesting", errs.E("no thing")),
			kind: errs.NotExist,
			want: false,
		},
		{
			err:  errs.E("nesting", errs.E("no thing inside")),
			kind: errs.Other,
			want: true,
		},
	}

	for _, tc := range testcases {
		got := errs.KindIs(tc.kind, tc.err)
		assert.Equal(t, tc.want, got, "KindIs: err(%s) want=(%s), got=(%s)", tc.err, tc.want, got)
	}
}

func TestMatch(t *testing.T) {
	user := errs.UserName("ardikabs")
	code := errs.Code("os_network")
	param := errs.Parameter("param")
	err := fmt.Errorf("network unreachable")

	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	want := errs.E(errs.IO, user, param, code, err)

	// Construct an error, one we pretend to have received from a test.
	err1 := errs.E(errs.IO, user, param, code, err)
	match := errs.Match(want, err1)
	assert.True(t, match, "Expect to be matched, but got mismatched")

	// Now one that's incorrect - wrong Kind.
	err2 := errs.E(errs.Database, user, param, code, err)
	match = errs.Match(want, err2)
	assert.False(t, match, "Expect to be mismatched, but matched")
}

func TestUnauthenticatedE(t *testing.T) {

	t.Run("default realm", func(t *testing.T) {
		err := errs.E(errs.Unauthenticated, errs.UserName("john@doe.com"))
		assert.NotNil(t, err)

		e, ok := err.(*errs.Error)
		assert.True(t, ok)

		assert.Equal(t, errs.DefaultRealm, e.Realm)
	})

	t.Run("given realm", func(t *testing.T) {
		err := errs.E(errs.Unauthenticated, errs.Realm("admin"), errs.UserName("john@doe.com"))
		assert.NotNil(t, err)

		e, ok := err.(*errs.Error)
		assert.True(t, ok)

		assert.Equal(t, errs.Realm("admin"), e.Realm)
	})
}
