package action

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
)

func TestScriptFactory_Name(t *testing.T) {
	f := ScriptFactory{}
	require.Equal(t, "script", f.Name())
}

func TestScript_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:    "When parameter 'script' is set then it executes the script",
			factory: ScriptFactory{},
			params: map[string]any{
				"script": "echo 'unittest' > test.txt",
			},
			wantFiles: map[string]string{"test.txt": "unittest\n"},
		},
		{
			name: "When parameter 'scriptFromFile' references a relative path then it reads the script relative to the task file and executes it",
			bootstrap: func(ctx context.Context) (string, context.Context) {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				scriptFilePath := "test.sh"
				f, err := os.Create(filepath.Join(tmpDir, scriptFilePath))
				require.NoError(t, err)
				_, err = f.WriteString("echo 'content from script file' > test.txt")
				require.NoError(t, err)
				_ = f.Close()
				return filepath.Join(tmpDir, "task.yaml"), ctx
			},
			factory: ScriptFactory{},
			params: map[string]any{
				"scriptFromFile": "test.sh",
			},
			wantFiles: map[string]string{"test.txt": "content from script file\n"},
		},
		{
			name: "When parameter 'scriptFromFile' references an absolute path then it reads and executes the script",
			bootstrap: func(ctx context.Context) (string, context.Context) {
				tmpDir := os.TempDir()
				scriptFilePath := "test.sh"
				f, err := os.Create(filepath.Join(tmpDir, scriptFilePath))
				require.NoError(t, err)
				_, err = f.WriteString("echo 'content from script file' > test.txt")
				require.NoError(t, err)
				_ = f.Close()
				return "", ctx
			},
			factory: ScriptFactory{},
			params: map[string]any{
				"scriptFromFile": filepath.Join(os.TempDir(), "test.sh"),
			},
			wantFiles: map[string]string{"test.txt": "content from script file\n"},
		},
		{
			name:    "When parameter 'scriptFromFile' references a file that odes not exist then it errors",
			factory: ScriptFactory{},
			params: map[string]any{
				"scriptFromFile": "/tmp/invalid.sh",
			},
			wantError: errors.New("read script from file: open /tmp/invalid.sh: no such file or directory"),
		},
		{
			name: "When the script contains a template variable then it renders the variable before execution",
			bootstrap: func(ctx context.Context) (string, context.Context) {
				data := template.FromContext(ctx)
				data.TaskName = "Template Variable Test"
				ctx = template.UpdateContext(ctx, data)
				return "", ctx
			},
			factory: ScriptFactory{},
			params: map[string]any{
				"script": "echo '{{.TaskName}}' > test.txt",
			},
			wantFiles: map[string]string{"test.txt": "Template Variable Test\n"},
		},
		{
			name:    "When the execution takes longer than 'timeout' then it errors",
			factory: ScriptFactory{},
			params: map[string]any{
				"script":  "sleep 30",
				"timeout": "1ms",
			},
			wantErrorApply: errors.New("execution of script took longer than 1ms\ncontext deadline exceeded"),
		},
		{
			name:      "When parameters 'script' and 'scriptFromFile' are missing then it errors",
			factory:   ScriptFactory{},
			params:    map[string]any{},
			wantError: errors.New("neither parameter `script` or `scriptFromFile` are set"),
		},
		{
			name:    "When both parameters 'script' and 'scriptFromFile' are set then it errors",
			factory: ScriptFactory{},
			params: map[string]any{
				"script":         "foo",
				"scriptFromFile": "bar",
			},
			wantError: errors.New("either parameter `script` or `scriptFromFile` need to be set, not both"),
		},
		{
			name: "When the script is executed then it has access to the env var TASK_DIR",
			bootstrap: func(ctx context.Context) (string, context.Context) {
				return "/tmp/task/unittest.yaml", ctx
			},
			factory: ScriptFactory{},
			params: map[string]any{
				"script": "echo $TASK_DIR > test.txt",
			},
			wantFiles: map[string]string{"test.txt": "/tmp/task\n"},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestScript_String(t *testing.T) {
	f := ScriptFactory{}
	a, err := f.Create(params.Params{"script": "echo 'test'"}, "")
	require.NoError(t, err)
	require.Equal(t, "script(shell=sh, timeout=10s)", a.String())
}
