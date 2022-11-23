package httpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// Doer is a standard http.Do interface, which provide flexibleness for the user
// to bring their own custom http client following with the standard net/http.Client
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Option represent the the request option
type Option func(*Request) error

// Request represent the http Request
type Request struct {
	ctx   context.Context
	httpc Doer
	url   *url.URL

	method      string
	header      http.Header
	queryParams url.Values
	body        io.Reader

	retryLimit int
	retryOn    []RetryOnFunc

	receiver interface{}

	unmarshalFunc UnmarshalFunc

	statusCodeHandlers StatusCodeHandlers
}

// NewRequest returns a new Request following with error
func NewRequest(client Doer, baseURL string, opts ...Option) (*Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return nil, ErrProtocolRequired
	}

	r := &Request{
		ctx:                context.Background(),
		httpc:              client,
		url:                u,
		header:             make(http.Header),
		queryParams:        make(url.Values),
		statusCodeHandlers: make(StatusCodeHandlers),
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Request) getURL() string {
	if r.url.RawQuery != "" {
		r.url.RawQuery = r.url.RawQuery + "&" + r.queryParams.Encode()
	} else {
		r.url.RawQuery = r.queryParams.Encode()
	}

	return r.url.String()
}

// Invoke do invoking the the Request to a given setup (URL, headers, body, parameters)
// and processing based on needs
func (r *Request) Invoke() error {
	req, err := http.NewRequestWithContext(r.ctx, r.method, r.getURL(), r.body)
	if err != nil {
		return err
	}
	req.Header = r.header

	var (
		retryCount int
		response   *http.Response
	)

doRequest:
	for ; retryCount <= r.retryLimit; retryCount++ {
		if response != nil {
			response.Body.Close()
		}

		var err error
		response, err = r.httpc.Do(req)
		if err != nil {
			return err
		}

		if len(r.retryOn) > 0 {
			for _, retryOn := range r.retryOn {
				if valid := retryOn(response.StatusCode); valid {
					continue doRequest
				}
			}
		}

		if handler, ok := r.statusCodeHandlers[response.StatusCode]; ok {
			if err := handler(response); err != nil {
				return err
			}

			return nil
		}

		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		contentType := response.Header.Get(HeaderContentType)
		if err := r.unmarshal(contentType, data, r.receiver); err != nil {
			return err
		}

		break
	}

	if retryCount > r.retryLimit {
		return ErrRetryExceeded
	}

	return nil
}

func (r *Request) unmarshal(contentType string, data []byte, out interface{}) error {
	// if user bring their own unmarshal solution
	// we assume it might be also customized how to interact
	// with the response receiver
	if r.unmarshalFunc != nil {
		return r.unmarshalFunc(contentType, data, out)
	}

	// if out is nil, will skip the unmarshal step
	// assuming user aware with that
	if out == nil {
		return nil
	}

	switch contentType {
	case MIMEApplicationJSON:
		fallthrough
	default:
		return json.Unmarshal(data, out)
	}
}
