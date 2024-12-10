package action

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	sbtemplate "github.com/wndhydrnt/saturn-bot/pkg/template"
)

// ScriptFactory initializes a new script action.
type ScriptFactory struct{}

// Create implements Factory.
func (f ScriptFactory) Create(params params.Params, taskPath string) (Action, error) {
	script, err := params.String("script", "")
	if err != nil {
		return nil, err
	}

	scriptFromFile, err := params.String("scriptFromFile", "")
	if err != nil {
		return nil, err
	}

	shell, err := params.String("shell", "sh")
	if err != nil {
		return nil, err
	}

	timeout, err := params.Duration("timeout", 10*time.Second)
	if err != nil {
		return nil, err
	}

	if script == "" && scriptFromFile == "" {
		return nil, fmt.Errorf("neither parameter `script` or `scriptFromFile` are set")
	}

	if script != "" && scriptFromFile != "" {
		return nil, fmt.Errorf("either parameter `script` or `scriptFromFile` need to be set, not both")
	}

	taskDirAbs, err := filepath.Abs(filepath.Dir(taskPath))
	if err != nil {
		return nil, fmt.Errorf("turn path of task directory into absolute path: %w", err)
	}

	var scriptContent string
	if scriptFromFile != "" {
		var path string
		if filepath.IsAbs(scriptFromFile) {
			path = scriptFromFile
		} else {
			path = filepath.Join(taskDirAbs, scriptFromFile)
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read script from file: %w", err)
		}

		scriptContent = string(b)
	} else {
		scriptContent = script
	}

	tpl, err := template.New("").Parse(scriptContent)
	if err != nil {
		return nil, fmt.Errorf("parse script as template: %w", err)
	}

	return &scriptAction{
		scriptTpl: tpl,
		shell:     shell,
		taskDir:   taskDirAbs,
		timeout:   timeout,
	}, nil
}

// Name implements Factory.
func (f ScriptFactory) Name() string {
	return "script"
}

type scriptAction struct {
	scriptTpl *template.Template
	shell     string
	taskDir   string
	timeout   time.Duration
}

// Apply implements Action.
func (a *scriptAction) Apply(ctx context.Context) error {
	scriptFile, err := os.CreateTemp("", "*.sh")
	if err != nil {
		return fmt.Errorf("create temporary script file: %w", err)
	}

	defer scriptFile.Close()
	templateData := sbtemplate.FromContext(ctx)
	err = a.scriptTpl.Execute(scriptFile, templateData)
	if err != nil {
		return fmt.Errorf("render script: %w", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Command(a.shell, scriptFile.Name()) // #nosec G204
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	env := cmd.Environ()
	env = append(env, "TASK_DIR="+a.taskDir)
	cmd.Env = env
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), a.timeout)
	defer cmdCancel()
	errChan := make(chan error)
	go func() {
		log.Log().Debugf("Executing script action '%s %s'", a.shell, scriptFile.Name())
		err := cmd.Run()
		errChan <- err
	}()
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("%w\nstdout: %s\nstderr:%s", err, stdout.String(), stderr.String())
		}

		return nil
	case <-cmdCtx.Done():
		var killErr error
		if cmd.Process != nil {
			killErr = cmd.Process.Kill()
		}
		return errors.Join(fmt.Errorf("execution of script took longer than %s", a.timeout), cmdCtx.Err(), killErr)
	}
}

// String implements Action.
func (a *scriptAction) String() string {
	return fmt.Sprintf("script(shell=%s, timeout=%s)", a.shell, a.timeout)
}
