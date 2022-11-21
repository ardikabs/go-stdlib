package errs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ardikabs/golib/pkg/errs"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHTTPStatusCodeFromError(t *testing.T) {
	type args struct {
		k errs.Kind
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"Validation", args{k: errs.Validation}, http.StatusBadRequest},
		{"Exist", args{k: errs.Exist}, http.StatusConflict},
		{"NotExist", args{k: errs.NotExist}, http.StatusNotFound},
		{"Invalid", args{k: errs.Invalid}, http.StatusNotAcceptable},
		{"Unauthenticated", args{k: errs.Unauthenticated}, http.StatusUnauthorized},
		{"Unauthorized", args{k: errs.Unauthorized}, http.StatusForbidden},
		{"Other", args{k: errs.Other}, http.StatusInternalServerError},
		{"Internal", args{k: errs.Internal}, http.StatusInternalServerError},
		{"Database", args{k: errs.Database}, http.StatusInternalServerError},
		{"Private", args{k: errs.Private}, http.StatusInternalServerError},
		{"Unidentified", args{k: errs.Kind(128)}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errs.HTTPStatusCodeFromError(errs.E(tt.args.k))
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("Unknown", func(t *testing.T) {
		got := errs.HTTPStatusCodeFromError(fmt.Errorf("unknown error"))
		assert.Equal(t, http.StatusNotImplemented, got)
	})
}

func TestHTTPErrorHandler_StatusCode(t *testing.T) {
	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	l := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	tests := []struct {
		name string
		args args
		want int
	}{
		{"nil error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: nil,
		}, http.StatusInternalServerError},
		{"unknown error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: fmt.Errorf("example of unknown error"),
		}, http.StatusNotImplemented},
		{"undefined error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Exist),
		}, http.StatusInternalServerError},
		{"unidentified error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Kind(128)),
		}, http.StatusInternalServerError},
		{"Validation", args{
			w: httptest.NewRecorder(),
			l: l,
			err: errs.E(errs.Validation, errs.ValidationErrors{
				errs.E(errs.Parameter("key")),
				errs.E(errs.Parameter("last_name")),
			}),
		}, http.StatusBadRequest},
		{"Validation without ValidationError", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Validation, errs.E("invalid error")),
		}, http.StatusInternalServerError},
		{"Unauthenticated", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Unauthenticated, errs.UserName("john@doe.com"), "unauthenticated user"),
		}, http.StatusUnauthorized},
		{"Unauthorized", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Unauthorized, errs.UserName("john@doe.com"), "unauthorized access"),
		}, http.StatusForbidden},
		{"common error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.NotExist, errs.Code("resource_not_exist"), "resource is not exist"),
		}, http.StatusNotFound},
		{"common error for internal", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Internal, errs.Code("internal"), "internal"),
		}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs.HTTPErrorHandler(tt.args.w, tt.args.l, tt.args.err)
			got := tt.args.w.Result().StatusCode
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHTTPErrorHandler_Body(t *testing.T) {
	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	l := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	tests := []struct {
		name string
		args args
		want string
	}{
		{"nil error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: nil,
		}, ""},
		{"unknown error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: fmt.Errorf("example of unknown error"),
		}, `{"error":{"code":"unknown_error","message":"unknown error - please contact support"}}`},
		{"undefined error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Exist),
		}, ""},
		{"Validation", args{
			w: httptest.NewRecorder(),
			l: l,
			err: errs.E(errs.Validation, errs.ValidationErrors{
				errs.E(errs.Parameter("key"), "bad format"),
				errs.E(errs.Parameter("last_name"), "bad format"),
			}),
		}, `{"errors":[{"param":"key","message":"bad format"},{"param":"last_name","message":"bad format"}]}`},
		{"Validation with unexpected error", args{
			w: httptest.NewRecorder(),
			l: l,
			err: errs.E(errs.Validation, errs.ValidationErrors{
				fmt.Errorf("unexpected error"),
				errs.E(errs.Parameter("key"), "bad format"),
			}),
		}, `{"errors":[{"param":"key","message":"bad format"}]}`},
		{"Unauthenticated", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Unauthenticated, errs.UserName("john@doe.com"), "unauthenticated user"),
		}, ""},
		{"Unauthorized", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Unauthorized, errs.UserName("john@doe.com"), "unauthorized access"),
		}, ""},
		{"common error", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.NotExist, errs.Code("product_not_exist"), "the product resource for id=14 is not exist"),
		}, `{"error":{"kind":"resource_does_not_exist","code":"product_not_exist","message":"the product resource for id=14 is not exist"}}`},
		{"common error for internal", args{
			w:   httptest.NewRecorder(),
			l:   l,
			err: errs.E(errs.Internal, "internal"),
		}, `{"error":{"kind":"internal_error","message":"internal server error"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs.HTTPErrorHandler(tt.args.w, tt.args.l, tt.args.err)
			got := tt.args.w.Body.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
