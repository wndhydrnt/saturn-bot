package action

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ExecFactory struct{}

func (f ExecFactory) Create(params map[string]any, taskPath string) (Action, error) {
	var args []string
	if params["args"] != nil {
		argsGeneric, ok := params["args"].([]any)
		if !ok {
			return nil, fmt.Errorf("parameter `args` is of type %T not list", params["args"])
		}

		for idx, argGeneric := range argsGeneric {
			arg, ok := argGeneric.(string)
			if !ok {
				return nil, fmt.Errorf("item `args[%d]` is of type %T not string", idx, argGeneric)
			}

			args = append(args, arg)
		}
	}

	if params["command"] == nil {
		return nil, fmt.Errorf("required parameter `command` not set")
	}
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `command` is of type %T not string", params["command"])
	}

	// Try to resolve relative path if command is not a name like "terraform"
	// or not an absolute path like "/usr/bin/terraform".
	if command != filepath.Base(command) && !filepath.IsAbs(command) {
		taskDir := filepath.Dir(taskPath)
		commandAbs, err := filepath.Abs(filepath.Join(taskDir, command))
		if err != nil {
			return nil, fmt.Errorf("make path to command absolute: %w", err)
		}

		command = commandAbs
	}

	var timeout time.Duration
	if params["timeout"] == nil {
		timeout = 2 * time.Minute
	} else {
		timeoutStr, ok := params["timeout"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `timeout` is of type %T not string", params["timeout"])
		}

		var parseErr error
		timeout, parseErr = time.ParseDuration(timeoutStr)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse parameter `timeout`: %w", parseErr)
		}
	}

	return &execAction{
		args:    args,
		name:    command,
		timeout: timeout,
	}, nil
}

func (f ExecFactory) Name() string {
	return "exec"
}

type execAction struct {
	args    []string
	name    string
	timeout time.Duration
}

func (a *execAction) Apply(_ context.Context) error {
	la := &logAdapter{name: filepath.Base(a.name)}
	cmd := exec.Command(a.name, a.args...) // #nosec G204 -- users can pass arbitrary values here
	cmd.Stdout = la
	cmd.Stderr = la
	errChan := make(chan error)
	go func() {
		err := cmd.Run()
		errChan <- err
	}()
	timer := time.NewTimer(a.timeout)
	select {
	case err := <-errChan:
		if !timer.Stop() {
			// Drain the channel
			<-timer.C
		}
		return err
	case <-timer.C:
		err := cmd.Process.Kill()
		return errors.Join(fmt.Errorf("command timed out"), err)
	}
}

func (a *execAction) String() string {
	args := strings.Join(a.args, ",")
	return fmt.Sprintf("exec(args=%s, command=%s, timeout=%s)", args, a.name, a.timeout)
}

type logAdapter struct {
	name string
}

func (la *logAdapter) Write(p []byte) (n int, err error) {
	slog.Debug(strings.TrimSpace(string(p)), "command", la.name)
	return len(p), nil
}
