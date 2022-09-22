package errs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type ValidationErrors []error

func (v *ValidationErrors) Append(args ...interface{}) {
	*v = append(*v, E(args...))
}

func (v ValidationErrors) Error() string {
	buff := bytes.NewBufferString("")

	for i := 0; i < len(v); i++ {

		if err, ok := v[i].(*Error); ok {
			buff.WriteString(fmt.Sprintf("%s: %s", err.Param, err.Error()))
			buff.WriteString("\n")
		}
	}

	return strings.TrimSpace(buff.String())
}

// HTTPErrResponse is used as the Response Body
type HTTPErrResponse struct {
	Error  *ServiceError  `json:"error,omitempty"`
	Errors []ServiceError `json:"errors,omitempty"`
}

// ServiceError has fields for Service errors. All fields with no data will
// be omitted
type ServiceError struct {
	Kind    string `json:"kind,omitempty"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message,omitempty"`
}

// HTTPErrorHandler is a pre-defined http error handler, it will translate given error structured response
// it also support to log given error
func HTTPErrorHandler(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	if err == nil {
		lgr.Error().
			Stack().
			Int("HTTP Error StatusCode", http.StatusInternalServerError).
			Msg("nil error - no response body sent")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var e *Error
	if errors.As(err, &e) {
		switch e.Kind {
		case Validation:
			validationErrHandler(w, lgr, e)
			return
		case Unauthenticated:
			unauthenticatedErrHandler(w, lgr, e)
			return
		case Unauthorized:
			unauthorizedErrHandler(w, lgr, e)
			return
		default:
			commonErrHandler(w, lgr, e)
			return
		}
	}

	unknownErrHandler(w, lgr, err)
}

func commonErrHandler(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	if e.isZero() {
		lgr.Error().Stack().
			Str("kind", string(e.Kind)).
			Msg("empty error")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lgr.Error().Stack().Err(e.Err).
		Str("kind", string(e.Kind)).
		Str("username", string(e.User)).
		Str("parameter", string(e.Param)).
		Str("code", string(e.Code)).
		Msg("common error")

	var errResponse HTTPErrResponse
	switch e.Kind {
	case Internal, Database, IO:
		errResponse = HTTPErrResponse{
			Error: &ServiceError{
				Kind:    string(e.Kind),
				Message: "internal server error",
			},
		}
	default:
		errResponse = HTTPErrResponse{
			Error: &ServiceError{
				Kind:    string(e.Kind),
				Code:    string(e.Code),
				Param:   string(e.Param),
				Message: e.Error(),
			},
		}
	}

	errJSON, _ := json.Marshal(errResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(errJSON)
	w.WriteHeader(HTTPStatusCodeFromError(e))
}

func validationErrHandler(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	verr, ok := e.Err.(ValidationErrors)
	if !ok {
		lgr.Error().Stack().Msg("validation error not having appropriate error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lgr.Error().
		Stack().
		Err(e.Err).
		Int("fields", len(verr)).
		Msg("input validation error")

	var errFields []ServiceError
	for _, err := range verr {
		ie, ok := err.(*Error)
		if !ok {
			continue
		}

		errFields = append(errFields, ServiceError{
			Code:    string(ie.Code),
			Param:   string(ie.Param),
			Message: ie.Error(),
		})
	}

	var errResponse = HTTPErrResponse{
		Errors: errFields,
	}

	errJSON, _ := json.Marshal(errResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(errJSON)
	w.WriteHeader(HTTPStatusCodeFromError(e))
}

func unauthenticatedErrHandler(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	lgr.Error().
		Stack().
		Err(e.Err).
		Str("realm", string(e.Realm)).
		Str("user", string(e.User)).
		Msg("unauthenticated request")

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, e.Realm))
	w.WriteHeader(HTTPStatusCodeFromError(e))
}

func unauthorizedErrHandler(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	lgr.Error().
		Stack().
		Err(e.Err).
		Str("realm", string(e.Realm)).
		Str("user", string(e.User)).
		Msg("unauthorized request")

	w.WriteHeader(HTTPStatusCodeFromError(e))
}

func unknownErrHandler(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	errResponse := HTTPErrResponse{
		Error: &ServiceError{
			Code:    "unknown_error",
			Message: "unknown error - please contact support",
		},
	}

	lgr.Error().Stack().Err(err).Msg("unknown error")

	errJSON, _ := json.Marshal(errResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(errJSON)
	w.WriteHeader(http.StatusInternalServerError)
}

// HTTPStatusCodeFromError translate error to an http status code
func HTTPStatusCodeFromError(err error) int {

	var e *Error
	if !errors.As(err, &e) {
		return http.StatusInternalServerError
	}

	switch e.Kind {
	case Validation:
		return http.StatusBadRequest
	case NotExist:
		return http.StatusNotFound
	case Invalid, InvalidRequest:
		return http.StatusNotAcceptable
	case Exist:
		return http.StatusConflict
	case Unauthenticated:
		return http.StatusUnauthorized
	case Unauthorized:
		return http.StatusForbidden
	case Other, IO, Internal, Private, Database:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
