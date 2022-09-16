package httpc_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardikabs/golib/pkg/httpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestSuccess(t *testing.T) {
	req, err := httpc.NewRequest(http.DefaultClient, "http://localhost.local")
	assert.NotNil(t, req)
	assert.Nil(t, err)
}

func TestNewRequestError(t *testing.T) {
	fakeBaseURL := "http://localhost.local"

	t.Run("bad http host", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, "localhost.local")
		assert.Nil(t, req)
		assert.NotNil(t, err)
		assert.Equal(t, httpc.ErrProtocolRequired, err)
	})

	t.Run("unparsed http base url", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, "postgres://user:abc{DEf1=ghi@example.com:5432/db?sslmode=require")
		assert.Nil(t, req)
		assert.NotNil(t, err)
	})

	t.Run("bad request payload", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, fakeBaseURL,
			httpc.WithRequestPayload(httpc.PayloadJSON, make(chan int)),
		)
		assert.NotNil(t, err)
		assert.Nil(t, req)
	})

	t.Run("response receiver couldn't nil", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, fakeBaseURL,
			httpc.WithResponseReceiver(nil),
		)
		assert.NotNil(t, err)
		assert.Nil(t, req)
	})

	t.Run("unknown retry on type", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, fakeBaseURL,
			httpc.WithRetryOn(0),
		)
		assert.NotNil(t, err)
		assert.Nil(t, req)
	})

	t.Run("custom status code handler nil", func(t *testing.T) {
		req, err := httpc.NewRequest(http.DefaultClient, fakeBaseURL,
			httpc.WithCustomHandler(http.StatusOK, nil),
		)
		assert.NotNil(t, err)
		assert.Nil(t, req)
	})
}

func TestInvokeSimple(t *testing.T) {
	expectedMethod := http.MethodGet
	expectedPath := "/api/v1/users"
	expectedQueryParamKey := "q"
	expectedQueryParamValue := "param1"
	expectedHeaderKey := "x-api-key"
	expectedHeaderValue := "simple-value"

	t.Run("simple", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, expectedHeaderValue, r.Header.Get(expectedHeaderKey))
			assert.Equal(t, expectedQueryParamValue, r.URL.Query().Get(expectedQueryParamKey))
		}))
		defer ts.Close()

		req, err := httpc.NewRequest(ts.Client(), ts.URL,
			httpc.WithMethod(expectedMethod),
			httpc.WithPath(expectedPath),
			httpc.WithQueryParam(expectedQueryParamKey, expectedQueryParamValue),
			httpc.WithHeader(expectedHeaderKey, expectedHeaderValue),
		)

		assert.NotNil(t, req)
		assert.Nil(t, err)

		err = req.Invoke()
		assert.Nil(t, err)
	})

	t.Run("simple with predetermined query params", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)
			assert.Equal(t, expectedHeaderValue, r.Header.Get(expectedHeaderKey))
			assert.Equal(t, expectedQueryParamValue, r.URL.Query().Get(expectedQueryParamKey))
			assert.Equal(t, "descending", r.URL.Query().Get("age"))
		}))
		defer ts.Close()

		url := fmt.Sprintf("%s?age=descending", ts.URL)

		req, err := httpc.NewRequest(ts.Client(), url,
			httpc.WithMethod(expectedMethod),
			httpc.WithPath(expectedPath),
			httpc.WithQueryParam(expectedQueryParamKey, expectedQueryParamValue),
			httpc.WithHeader(expectedHeaderKey, expectedHeaderValue),
		)

		assert.NotNil(t, req)
		assert.Nil(t, err)

		err = req.Invoke()
		assert.Nil(t, err)
	})
}

func TestInvokeWithPayloadAndReceiver(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	type response struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	expectedMethod := http.MethodPost
	expectedPath := "/api/v1/users"
	expectedPayload := payload{Name: "go-client"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedMethod, r.Method)
		assert.Equal(t, expectedPath, r.URL.Path)
		assert.Equal(t, httpc.MIMEApplicationJSON, r.Header.Get(httpc.HeaderContentType))

		assert.NotNil(t, r.Body)

		rbody, err := io.ReadAll(r.Body)
		require.NoError(t, err, "should not have failed to extract request body")

		expectedPayloadByte, err := json.Marshal(expectedPayload)
		require.NoError(t, err, "should not have failed to marshal the struct")

		assert.Equal(t, string(expectedPayloadByte), string(rbody))
		w.Header().Set(httpc.HeaderContentType, httpc.MIMEApplicationJSON)
		w.Write([]byte(`{"id": 1, "name": "fake name"}`))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	var resp response
	req, err := httpc.NewRequest(ts.Client(), ts.URL,
		httpc.WithMethod(expectedMethod),
		httpc.WithPath(expectedPath),
		httpc.WithRequestPayload(httpc.PayloadJSON, expectedPayload),
		httpc.WithResponseReceiver(&resp),
	)

	assert.NotNil(t, req)
	assert.Nil(t, err)

	err = req.Invoke()
	assert.Nil(t, err)

	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "fake name", resp.Name)
}

func TestInvokeWithRetry(t *testing.T) {
	expectedMethod := http.MethodGet
	expectedPath := "/api/v1/users"

	t.Run("retry on non 2xx", func(t *testing.T) {
		count := 0
		noOfRetries := 3
		expectedNoOfCalls := noOfRetries + 1

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)

			w.WriteHeader(http.StatusBadRequest)
			count++
		}))
		defer ts.Close()

		req, err := httpc.NewRequest(ts.Client(), ts.URL,
			httpc.WithMethod(expectedMethod),
			httpc.WithPath(expectedPath),
			httpc.WithRetryOn(httpc.RetryOnNon2xx),
			httpc.WithRetryLimit(noOfRetries),
		)

		assert.NotNil(t, req)
		assert.Nil(t, err)

		err = req.Invoke()
		assert.ErrorIs(t, err, httpc.ErrRetryExceeded)
		assert.Equal(t, expectedNoOfCalls, count)
	})

	t.Run("retry on 4xx", func(t *testing.T) {
		count := 0
		noOfRetries := 3
		expectedNoOfCalls := noOfRetries + 1

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)

			count++
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		req, err := httpc.NewRequest(ts.Client(), ts.URL,
			httpc.WithMethod(expectedMethod),
			httpc.WithPath(expectedPath),
			httpc.WithRetryOn(httpc.RetryOn4xx),
			httpc.WithRetryLimit(noOfRetries),
		)

		assert.NotNil(t, req)
		assert.Nil(t, err)

		err = req.Invoke()
		assert.ErrorIs(t, err, httpc.ErrRetryExceeded)
		assert.Equal(t, expectedNoOfCalls, count)
	})

	t.Run("retry on gateway error", func(t *testing.T) {
		count := 0
		noOfRetries := 3
		expectedNoOfCalls := noOfRetries + 1

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)

			count++

			w.WriteHeader(502)
		}))
		defer ts.Close()

		req, err := httpc.NewRequest(ts.Client(), ts.URL,
			httpc.WithMethod(expectedMethod),
			httpc.WithPath(expectedPath),
			httpc.WithRetryOn(httpc.RetryOnGatewayErr),
			httpc.WithRetryLimit(noOfRetries),
		)

		assert.NotNil(t, req)
		assert.Nil(t, err)

		err = req.Invoke()
		assert.ErrorIs(t, err, httpc.ErrRetryExceeded)
		assert.Equal(t, expectedNoOfCalls, count)
	})
}

func TestInvokeWithUnmarshaller(t *testing.T) {
	expectedMethod := http.MethodPost
	expectedPath := "/api/v1/users"
	expectedResponseByte := []byte(`{ "name": "fake go-client"}`)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedMethod, r.Method)
		assert.Equal(t, expectedPath, r.URL.Path)

		w.Header().Set(httpc.HeaderContentType, httpc.MIMEApplicationJSON)
		w.Write(expectedResponseByte)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	req, err := httpc.NewRequest(ts.Client(), ts.URL,
		httpc.WithMethod(expectedMethod),
		httpc.WithPath(expectedPath),
		httpc.WithUnmarshaler(func(contentType string, data []byte, out interface{}) error {
			assert.Equal(t, string(expectedResponseByte), string(data))
			assert.Equal(t, httpc.MIMEApplicationJSON, contentType)
			return nil
		}),
	)

	assert.NotNil(t, req)
	assert.Nil(t, err)

	err = req.Invoke()
	assert.Nil(t, err)
}

func TestInvokeWithCustomHandler(t *testing.T) {
	expectedMethod := http.MethodPost
	expectedPath := "/api/v1/users"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedMethod, r.Method)
		assert.Equal(t, expectedPath, r.URL.Path)

		w.Header().Set(httpc.HeaderContentType, httpc.MIMEApplicationJSON)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	req, err := httpc.NewRequest(ts.Client(), ts.URL,
		httpc.WithMethod(expectedMethod),
		httpc.WithPath(expectedPath),
		httpc.WithCustomHandler(http.StatusAccepted, func(r *http.Response) error {
			assert.Equal(t, http.StatusAccepted, r.StatusCode)
			assert.Equal(t, httpc.MIMEApplicationJSON, r.Header.Get(httpc.HeaderContentType))
			return nil
		}),
	)

	assert.NotNil(t, req)
	assert.Nil(t, err)

	err = req.Invoke()
	assert.Nil(t, err)
}
