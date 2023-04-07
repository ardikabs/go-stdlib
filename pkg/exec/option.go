package exec

import (
	"fmt"
	"strings"
)

type Option func(*Task) error

func WithExecutor(exec TaskExecutor) Option {
	return func(t *Task) error {
		t.exec = exec
		return nil
	}
}

func WithDirectory(dir string) Option {
	return func(t *Task) error {
		t.cwd = dir
		return nil
	}
}

func WithArgs(args ...string) Option {
	return func(t *Task) error {
		t.args = args
		return nil
	}
}

func WithEnv(envs ...string) Option {
	return func(t *Task) error {
		if len(envs) == 0 {
			return nil
		}

		for _, env := range envs {
			if len(strings.Split(env, "=")) == 1 {
				return fmt.Errorf("environment variable has an invalid format %s, correct format (key=value)", env)
			}
		}

		t.env = envs
		return nil
	}
}

func WithEnableStreamIO() Option {
	return func(t *Task) error {
		t.streamIO = true
		return nil
	}
}

func WithShell(shell string) Option {
	return func(t *Task) error {

		if shell == "" {
			return fmt.Errorf("shell couldn't be empty")
		}

		t.shellExec = shell
		t.shellMode = true
		return nil
	}
}

func WithDebug() Option {
	return func(t *Task) error {
		t.debug = true
		return nil
	}
}
