package httpc

import (
	"bytes"
	"context"
	"encoding/json"

	"fmt"

	"github.com/ardikabs/golib/pkg/tool"
)

// WithContext set the context for the http request operation
func WithContext(ctx context.Context) Option {
	return func(req *Request) error {
		req.ctx = ctx
		return nil
	}
}

// WithPath set the URL Path
func WithPath(path string) Option {
	return func(req *Request) error {
		req.url.Path = path
		return nil
	}
}

// WithQueryParam set the URL Query Params
func WithQueryParam(key, value string) Option {
	return func(req *Request) error {
		req.queryParams.Add(key, value)
		return nil
	}
}

// WithMethod set the HTTP Method
func WithMethod(method string) Option {
	return func(req *Request) error {
		req.method = method
		return nil
	}
}

// WithHeader set the HTTP Request Header
func WithHeader(key, value string) Option {
	return func(req *Request) error {
		req.header.Set(key, value)
		return nil
	}
}

// WithRequestPayload set the given user payload based on payload kind
// in short, this operation will do marshalling the user payload
func WithRequestPayload(kind PayloadKind, payload interface{}) Option {
	return func(req *Request) error {
		switch kind {
		case PayloadJSON:
			buf, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			req.body = bytes.NewBuffer(buf)
			if ct := req.header.Get(HeaderContentType); ct == "" {
				req.header.Set(HeaderContentType, MIMEApplicationJSON)
			}
		}
		return nil
	}
}

// WithResponseReceiver set the receiver for the response Body on Request getting invoked
func WithResponseReceiver(receiver interface{}) Option {
	return func(req *Request) error {
		if receiver == nil {
			return fmt.Errorf("response receiver MUST not be nil")
		}

		req.receiver = receiver
		return nil
	}
}

// WithUnmarshaler set a custom unmarshaller solution
func WithUnmarshaler(fn UnmarshalFunc) Option {
	return func(req *Request) error {
		req.unmarshalFunc = fn
		return nil
	}
}

// WithCustomHandler set custom handler for a given HTTP status code
func WithCustomHandler(code int, fn StatusCodeHandleFunc) Option {
	return func(req *Request) error {
		if fn == nil {
			return fmt.Errorf("status code handle func MUST not be nil")
		}

		req.statusCodeHandlers.Set(code, fn)
		return nil
	}

}

type (
	RetryOnKind uint
	RetryOnFunc func(code int) bool
)

const (
	// RetryOnNon2xx represent a retry event when the HTTP response status code is not part of 2xx
	RetryOnNon2xx RetryOnKind = iota + 1
	// RetryOn4xx represent a retry event when the HTTP response status code greater than 400
	RetryOn4xx
	// RetryOnGatewayErr represent a retry event when the HTTP response status code is one of 502,503,504 (collection of a gateway error condition)
	RetryOnGatewayErr
)

// WithRetryOn set a retryOn condition on invoking the HTTP request
func WithRetryOn(retryOn RetryOnKind) Option {
	return func(req *Request) error {

		var fn RetryOnFunc

		switch retryOn {
		case RetryOnNon2xx:
			fn = func(code int) bool {
				return code < 200 || code >= 299
			}
		case RetryOn4xx:
			fn = func(code int) bool {
				return code >= 400
			}
		case RetryOnGatewayErr:
			fn = func(code int) bool {
				return tool.In(code, 502, 503, 504)
			}
		default:
			return fmt.Errorf("unknown retry on type")
		}

		req.retryOn = append(req.retryOn, fn)
		return nil
	}
}

// WithRetryLimit set a retry limit, this is a counterpart within WithRetryOn
func WithRetryLimit(limit int) Option {
	return func(req *Request) error {
		req.retryLimit = limit
		return nil
	}
}
