package httpclient

import "fmt"

var (
	// ErrRetryExceeded
	ErrRetryExceeded = fmt.Errorf("http retry exceeded the limit")

	// ErrProtocolRequired
	ErrProtocolRequired = fmt.Errorf("invalid http host, protocol/scheme is required")
)
