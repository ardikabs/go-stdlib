package errs_test

import (
	"testing"

	"github.com/ardikabs/golib/errs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrsNoArgs(t *testing.T) {
	assert.Panics(t, func() {
		_ = errs.E()
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
	err := errors.New("network unreachable")

	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	want := errs.E(errs.IO, user, err)

	// Construct an error, one we pretend to have received from a test.
	err1 := errs.E(errs.IO, user, err)
	match := errs.Match(want, err1)
	assert.True(t, match, "Expect to be matched, but got mismatched")

	// Now one that's incorrect - wrong Kind.
	err2 := errs.E(errs.Database, user, err)
	match = errs.Match(want, err2)
	assert.False(t, match, "Expect to be mismatched, but matched")
}
