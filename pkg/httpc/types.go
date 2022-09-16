package httpc

import "net/http"

// PayloadKind is an identifier for identify the kind of user payload
type PayloadKind uint

const (
	// PayloadJSON is an indentifier for JSON payload
	PayloadJSON PayloadKind = iota + 1
)

// StatusCodeHandleFunc is a func type with http.Response as parameter
type StatusCodeHandleFunc func(*http.Response) error

// StatusCodeHandlers is a pair of http status code and the handler
type StatusCodeHandlers map[int]StatusCodeHandleFunc

// Set will define the handler for a given http status code
// if a given http status code already exists in the map, the next operation would be overridden the previous handler for the http status code
func (s StatusCodeHandlers) Set(code int, handler StatusCodeHandleFunc) {
	s[code] = handler
}

// UnmarshalFunc intended for user if they want bring their own unmarshaller solution
type UnmarshalFunc func(contentType string, data []byte, out interface{}) error
