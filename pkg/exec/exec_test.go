package exec_test

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/ardikabs/golib/pkg/exec"
	"github.com/stretchr/testify/assert"
)

func TestShellProcessSuccess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	os.Exit(0)
}

func TestNewExec(t *testing.T) {

	t.Run("ideal initiation", func(t *testing.T) {
		tc, err := NewExec("cmd")

		assert.NotNil(t, tc)
		assert.NoError(t, err)

	})

	t.Run("init cmd with additional options", func(t *testing.T) {
		tc, err := NewExec("cmd",
			WithArgs("-c", "some"),
			WithDirectory("/tmp"),
			WithShell("/bin/sh"),
			WithEnableStreamIO(),
			WithDebug())

		assert.NotNil(t, tc)
		assert.NoError(t, err)
	})

	t.Run("init cmd with invalid environment variable format", func(t *testing.T) {
		tc, err := NewExec("cmd",
			WithArgs("-c", "some"),
			WithDirectory("/tmp"),
			WithShell("/bin/sh"),
			WithEnv("bloblo=bleble", "keyvalue"),
			WithEnableStreamIO(),
			WithDebug())

		assert.Nil(t, tc)
		assert.Error(t, err)
	})
}

func TestExecute(t *testing.T) {
	fakeExecutor := func(command string, args ...string) *exec.Cmd {
		cs := []string{
			"-test.run=TestShellProcessSuccess",
			"--",
			command,
		}

		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
		return cmd
	}

	t.Run("ideal task execution", func(t *testing.T) {
		tc, err := NewExec("ping", WithExecutor(fakeExecutor))
		assert.NoError(t, err)
		assert.NotNil(t, tc)

		result := tc.Execute()
		assert.Equal(t, "ping", result.Command)
	})

	t.Run("task execution with arguments and environment variables", func(t *testing.T) {
		tc, err := NewExec("ping",
			WithExecutor(fakeExecutor),
			WithArgs("-c", "1", "google.com"),
			WithEnv("KEY=VALUE", "KEY1=VALUE1"),
		)
		assert.NoError(t, err)
		assert.NotNil(t, tc)

		result := tc.Execute()
		assert.Equal(t, "ping", result.Command)
		assert.Len(t, result.Args, 3)
		assert.Len(t, result.Env, 2)
	})
}
