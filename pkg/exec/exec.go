package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	defaultFailedCode = 1
)

type TaskExecutor func(command string, args ...string) *exec.Cmd

type Task struct {
	exec TaskExecutor

	command string
	args    []string
	env     []string
	cwd     string

	streamIO bool
	debug    bool

	shellExec string
	shellMode bool
}

type Result struct {
	Command string
	Args    []string
	Env     []string

	Stdout   string
	Stderr   string
	ExitCode int
}

func NewExec(command string, opts ...Option) (*Task, error) {
	t := &Task{
		command: command,
		exec:    exec.Command,
	}

	for _, o := range opts {
		if err := o(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func MustExec(command string, opts ...Option) *Task {
	t, err := NewExec(command, opts...)
	if err != nil {
		panic(err)
	}

	return t
}

func (t *Task) Execute() (result Result) {
	log := log.With().
		Str("dir", t.cwd).
		Str("cmd", t.command).
		Str("args", strings.Join(t.args, " ")).
		Logger()

	var outbuf, errbuf bytes.Buffer

	if t.shellMode {
		if t.shellExec == "" {
			t.shellExec = "/bin/sh"
		}
		t.args = append([]string{"-c", t.command}, t.args...)
		t.command = t.shellExec

		log = log.With().Str("shell", t.shellExec).Logger()
	}

	cmd := t.exec(t.command, t.args...)
	cmd.Dir = t.cwd
	cmd.Stdin = os.Stdin

	if t.streamIO {
		cmd.Stdout = io.MultiWriter(os.Stdout, &outbuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &errbuf)
	} else {
		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf
	}

	if len(t.env) > 0 {
		log = log.With().
			Str("env", strings.Join(t.env, ",")).
			Logger()

		overrides := map[string]bool{}
		for _, env := range t.env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}

		for _, env := range os.Environ() {
			key := strings.Split(env, "=")[0]

			if _, ok := overrides[key]; !ok {
				cmd.Env = append(cmd.Env, env)
			}
		}
	}

	if t.debug {
		log.Debug().Msg("executing the command")
	}

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			// This will happen (in OSX) if `t.Command` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr

			if t.debug {
				log.Debug().Msg("could not get exit code for failed command, return with default failed exit code")
			}

			result.ExitCode = defaultFailedCode
			if result.Stderr == "" {
				result.Stderr = err.Error()
			}
		}
	}

	result.Command = t.command
	result.Args = t.args
	result.Env = t.env
	result.Stdout = outbuf.String()
	result.Stderr = errbuf.String()

	return
}
