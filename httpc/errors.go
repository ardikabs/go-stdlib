package httpc

import "fmt"

var (
	// ErrRetryExceeded
	ErrRetryExceeded = fmt.Errorf("http retry exceeded the limit")
)
