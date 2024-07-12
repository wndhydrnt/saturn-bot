package action

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const script = `#!/usr/bin/env bash

first="$1"
second="$2"
echo "Writing to file"
echo "$first $second" > exec.txt
`

func TestExec_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name: "When a command is defined then it executes the command",
			bootstrap: func() string {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				scriptFile := "test.sh"
				f, err := os.Create(filepath.Join(tmpDir, scriptFile))
				require.NoError(t, err)
				_, err = f.WriteString(script)
				require.NoError(t, err)
				_ = f.Chmod(0755)
				_ = f.Close()
				return filepath.Join(tmpDir, "task.yaml")
			},
			factory: ExecFactory{},
			params: map[string]any{
				"args":    []any{"exec", "test"},
				"command": "./test.sh",
				"timeout": "1m",
			},
			wantError: nil,
			wantFiles: map[string]string{"exec.txt": "exec test\n"},
		},
		{
			name: "When the command points to a name then it executes the command",
			bootstrap: func() string {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				return filepath.Join(tmpDir, "task.yaml")
			},
			factory: ExecFactory{},
			params: map[string]any{
				"args":    []any{"-c", "echo 'name test' >> name.txt"},
				"command": "sh",
				"timeout": "1m",
			},
			wantError: nil,
			wantFiles: map[string]string{"name.txt": "name test\n"},
		},
		{
			name: "When the command point to an absolute path then it executes the command",
			bootstrap: func() string {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				return filepath.Join(tmpDir, "task.yaml")
			},
			factory: ExecFactory{},
			params: map[string]any{
				"args":    []any{"-c", "echo 'abs test' >> abs.txt"},
				"command": "/bin/sh",
				"timeout": "1m",
			},
			wantError: nil,
			wantFiles: map[string]string{"abs.txt": "abs test\n"},
		},
		{
			name:    "When parameter args not a list then it errors",
			factory: ExecFactory{},
			params: map[string]any{
				"args":    "exec",
				"command": "./test.sh",
			},
			wantError: errors.New("parameter `args` is of type string not list"),
		},
		{
			name:    "When an item of parameter args is not a string then it errors",
			factory: ExecFactory{},
			params: map[string]any{
				"args":    []any{"exec", 2},
				"command": "./test.sh",
			},
			wantError: errors.New("item `args[1]` is of type int not string"),
		},
		{
			name:      "When required parameter command is not set then it errors",
			factory:   ExecFactory{},
			params:    map[string]any{},
			wantError: errors.New("required parameter `command` not set"),
		},
		{
			name:    "When parameter command is not a string then it errors",
			factory: ExecFactory{},
			params: map[string]any{
				"command": 5,
			},
			wantError: errors.New("parameter `command` is of type int not string"),
		},
		{
			name:    "When parameter timeout is not a Go duration then it errors",
			factory: ExecFactory{},
			params: map[string]any{
				"command": "./test.sh",
				"timeout": "abc",
			},
			wantError: errors.New("failed to parse parameter `timeout`: time: invalid duration \"abc\""),
		},
		{
			name:    "When parameter timeout is not a string then it errors",
			factory: ExecFactory{},
			params: map[string]any{
				"command": "./test.sh",
				"timeout": 60,
			},
			wantError: errors.New("parameter `timeout` is of type int not string"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
