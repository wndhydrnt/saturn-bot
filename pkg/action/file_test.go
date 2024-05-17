package action

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileCreate_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:    "When parameter `content` is set and the file does not exist then it creates the file",
			files:   map[string]string{},
			factory: FileCreateFactory{},
			params: map[string]string{
				"content": "abc\n",
				"path":    "test.txt",
			},
			wantError: nil,
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When parameter `contentFile` is set and the file does not exist then it creates the file",
			files: map[string]string{},
			bootstrap: func() string {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				contentFilePath := "content.txt"
				f, err := os.Create(filepath.Join(tmpDir, contentFilePath))
				require.NoError(t, err)
				_, err = f.WriteString("abc\n")
				require.NoError(t, err)
				_ = f.Close()
				return filepath.Join(tmpDir, "task.yaml")
			},
			factory: FileCreateFactory{},
			params: map[string]string{
				"contentFromFile": "content.txt",
				"path":            "test.txt",
			},
			wantError: nil,
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:    "When `overwrite=false` and the file exists then it does not update the file",
			files:   map[string]string{"test.txt": "abc\n"},
			factory: FileCreateFactory{},
			params: map[string]string{
				"content":   "def\n",
				"overwrite": "false",
				"path":      "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:    "When `overwrite=true` and the file exists then it updates the file",
			files:   map[string]string{"test.txt": "abc\n"},
			factory: FileCreateFactory{},
			params: map[string]string{
				"content":   "def\n",
				"overwrite": "true",
				"path":      "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:    "When `overwrite=true` and the file does not exist then it creates the file",
			files:   map[string]string{},
			factory: FileCreateFactory{},
			params: map[string]string{
				"content":   "def\n",
				"overwrite": "true",
				"path":      "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:    "When `path` contains directories and those directories do not exist then it creates the directories and the file",
			files:   map[string]string{},
			factory: FileCreateFactory{},
			params: map[string]string{
				"content": "abc\n",
				"path":    "unit/test/test.txt",
			},
			wantFiles: map[string]string{"unit/test/test.txt": "abc\n"},
		},
		{
			name:    "When parameter `path` is not set then it errors",
			factory: FileCreateFactory{},
			params: map[string]string{
				"content": "abc\n",
			},
			wantError: errors.New("required parameter `path` not set"),
		},
		{
			name:    "When parameters `content` and `contentFromFile` are not set then it errors",
			factory: FileCreateFactory{},
			params: map[string]string{
				"path": "test.txt",
			},
			wantError: errors.New("either parameter `content` or `contentFromFile` is required"),
		},
		{
			name:    "When parameters `content` and `contentFromFile` are both set then it errors",
			factory: FileCreateFactory{},
			params: map[string]string{
				"content":         "abc\n",
				"contentFromFile": "content.txt",
				"path":            "test.txt",
			},
			wantError: errors.New("parameters `content` and `contentFromFile` cannot be set at the same time"),
		},
		{
			name:    "When parameter `mode` is invalid then it errors",
			factory: FileCreateFactory{},
			params: map[string]string{
				"content": "abc\n",
				"mode":    "75r",
				"path":    "test.txt",
			},
			wantError: errors.New("parse value of parameter `mode`: strconv.ParseUint: parsing \"75r\": invalid syntax"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestFileCreate_Apply_FileMode(t *testing.T) {
	workDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	fac := FileCreateFactory{}
	a, err := fac.Create(map[string]string{
		"content": "echo Unit Test",
		"mode":    "755",
		"path":    "test.sh",
	}, "")
	require.NoError(t, err)
	err = inDirectory(workDir, func() error {
		return a.Apply(context.Background())
	})
	require.NoError(t, err)

	fi, err := os.Stat(filepath.Join(workDir, "test.sh"))
	require.NoError(t, err)
	assert.Equal(t, fs.FileMode(0755), fi.Mode())
}

func TestFileDelete_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:    "When the file exists then it deletes the file",
			files:   map[string]string{"test.txt": "abc\n"},
			factory: FileDeleteFactory{},
			params: map[string]string{
				"path": "test.txt",
			},
			wantFiles: map[string]string{},
		},
		{
			name:    "When the file does not exist then it does nothing",
			files:   map[string]string{},
			factory: FileDeleteFactory{},
			params: map[string]string{
				"path": "test.txt",
			},
			wantFiles: map[string]string{},
		},
		{
			name:      "When parameter `path` is not set then it errors",
			files:     map[string]string{},
			factory:   FileDeleteFactory{},
			params:    map[string]string{},
			wantError: errors.New("required parameter `path` not set"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
