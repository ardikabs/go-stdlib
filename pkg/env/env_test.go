package env_test

import (
	"os"
	"testing"
	"time"

	"github.com/ardikabs/go-stdlib/pkg/env"
	"github.com/stretchr/testify/require"
)

func TestLookup(t *testing.T) {
	t.Setenv("HTTP_ADDR", "0.0.0.0")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("DEBUG_MODE", "1")
	t.Setenv("HTTP_TIMEOUT", "30s")
	t.Setenv("HTTP_RESERVED_CONTENT_TYPES", "text/plain,application/json")
	t.Setenv("HTTP_RESERVED_STATUS_CODES", "401,429,503")
	defer os.Clearenv()

	gotString := env.Lookup("HTTP_ADDR", "127.0.0.1")
	require.Equal(t, "0.0.0.0", gotString)

	gotInt := env.Lookup("HTTP_PORT", 80)
	require.Equal(t, 8080, gotInt)

	gotBool := env.Lookup("DEBUG_MODE", false)
	require.Equal(t, true, gotBool)

	gotDuration := env.Lookup("HTTP_TIMEOUT", time.Duration(15*time.Second))
	require.Equal(t, time.Duration(30*time.Second), gotDuration)

	gotStringArr := env.Lookup("HTTP_RESERVED_CONTENT_TYPES", []string{"text/html"})
	require.Equal(t, []string{"text/plain", "application/json"}, gotStringArr)

	gotIntArr := env.Lookup("HTTP_RESERVED_STATUS_CODES", []int{500})
	require.Equal(t, []int{401, 429, 503}, gotIntArr)
}

func TestLookupDefaultValue(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("HTTP_PORT", "invalid-value")
	t.Setenv("DEBUG_MODE", "invalid-value")
	t.Setenv("HTTP_TIMEOUT", "invalid-value")
	t.Setenv("HTTP_RESERVED_CONTENT_TYPES", "")
	t.Setenv("HTTP_RESERVED_STATUS_CODES", "")
	t.Setenv("RANGE_OF_RETRY_BACKOFF_SECONDS", "10,20,30a")
	defer os.Clearenv()

	got := env.Lookup("AGENT_MODE", true)
	require.Equal(t, true, got)

	gotString := env.Lookup("HTTP_ADDR", "127.0.0.1")
	require.Equal(t, "127.0.0.1", gotString)

	gotInt := env.Lookup("HTTP_PORT", 80)
	require.Equal(t, 80, gotInt)

	gotBool := env.Lookup("DEBUG_MODE", false)
	require.Equal(t, false, gotBool)

	gotDuration := env.Lookup("HTTP_TIMEOUT", time.Duration(15*time.Second))
	require.Equal(t, time.Duration(15*time.Second), gotDuration)

	gotStringArr := env.Lookup("HTTP_RESERVED_CONTENT_TYPES", []string{"text/html"})
	require.Equal(t, []string{"text/html"}, gotStringArr)

	gotIntArr := env.Lookup("HTTP_RESERVED_STATUS_CODES", []int{500})
	require.Equal(t, []int{500}, gotIntArr)

	gotBadIntArr := env.Lookup("RANGE_OF_RETRY_BACKOFF_SECONDS", []int{1, 5, 10})
	require.Equal(t, []int{1, 5, 10}, gotBadIntArr)
}
