package httpc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardikabs/go-client/httpc"
	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	req, err := httpc.NewRequest(http.DefaultClient, "localhost.local")
	assert.NotNil(t, req)
	assert.Nil(t, err)
}

func TestInvoke(t *testing.T) {
	expectedMethod := http.MethodGet
	expectedURL := "http://localhost.local/api/v1/users?query=param"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Fatalf("Method want=%s, got=%s", expectedMethod, r.Method)
		}

		if r.URL.String() != expectedURL {
			t.Fatalf("URL want=%s, got=%s", expectedURL, r.URL.String())
		}
	}))
	defer ts.Close()

	req, err := httpc.NewRequest(ts.Client(), "http://localhost.local",
		httpc.WithMethod(http.MethodGet),
		httpc.WithPath("/api/v1/users"),
		httpc.WithQueryParam("query", "param"),
	)

	assert.NotNil(t, req)
	assert.Nil(t, err)

	err = req.Invoke()
	assert.Nil(t, err)
}
