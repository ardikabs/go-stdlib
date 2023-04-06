package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultFailedCode = 1
)

type Task struct {
	Command  string
	Args     []string
	Env      []string
	Cwd      string
	StreamIO bool
	Debug    bool

	ShellExec string
	ShellMode bool
}

type TaskResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (t *Task) Execute() (result TaskResult) {

	var outbuf, errbuf bytes.Buffer

	if t.ShellMode {
		if t.ShellExec == "" {
			t.ShellExec = "/bin/sh"
		}
		t.Args = append([]string{"-c", t.Command}, t.Args...)
		t.Command = t.ShellExec
	}

	cmd := exec.Command(t.Command, t.Args...)
	cmd.Stdin = os.Stdin
	cmd.Dir = t.Cwd

	if len(t.Env) > 0 {
		overrides := map[string]bool{}
		for _, env := range t.Env {
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

	if t.StreamIO {
		cmd.Stdout = io.MultiWriter(os.Stdout, &outbuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &errbuf)
	} else {
		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf
	}

	if t.Debug {
		fmt.Println("executing: ", t.Command, strings.Join(t.Args, " "))
	}

	err := cmd.Run()

	result.Stdout = outbuf.String()
	result.Stderr = errbuf.String()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			// This will happen (in OSX) if `t.Command` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr

			if t.Debug {
				fmt.Printf("could not get exit code for failed program: %v, %v", t.Command, strings.Join(t.Args, " "))
			}
			result.ExitCode = defaultFailedCode
			if result.Stderr == "" {
				result.Stderr = err.Error()
			}
		}
	}
	return
}
