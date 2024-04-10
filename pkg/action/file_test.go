package action

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile_Apply(t *testing.T) {
	type input struct {
		content   string
		mode      string
		overwrite *bool
		path      string
		state     string
	}

	testCases := []struct {
		name      string
		files     map[string]string
		input     input
		wantFiles map[string]string
	}{
		{
			name:      "When state=present and overwrite=false and the file does not exist then it creates the file",
			files:     map[string]string{},
			input:     input{content: "abc\n", path: "test.txt", state: "present", overwrite: ptr(false)},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:      "When state=present and overwrite=false and the file exists then it does not update the file",
			files:     map[string]string{"test.txt": "abc\n"},
			input:     input{content: "def\n", path: "test.txt", state: "present", overwrite: ptr(false)},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:      "When state=present and overwrite=true and the file exists then it updates the file",
			files:     map[string]string{"test.txt": "abc\n"},
			input:     input{content: "def\n", path: "test.txt", state: "present", overwrite: ptr(true)},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:      "When state=present and overwrite=true and the file does not exist then it creates the file",
			files:     map[string]string{"test.txt": "abc\n"},
			input:     input{content: "def\n", path: "test.txt", state: "present", overwrite: ptr(true)},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:      "When state=present and is not set explicitly and the file exists then it updates the file",
			files:     map[string]string{"test.txt": "abc\n"},
			input:     input{content: "def\n", path: "test.txt", state: "present"},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:      "When state=absent the file exists then it deletes the file",
			files:     map[string]string{"test.txt": "abc\n"},
			input:     input{path: "test.txt", state: "absent"},
			wantFiles: map[string]string{},
		},
		{
			name:      "When state=absent the file does not exist then it does nothing",
			files:     map[string]string{},
			input:     input{path: "test.txt", state: "absent"},
			wantFiles: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workDir, err := setupTestFiles(tc.files)
			require.NoError(t, err)

			rc := io.NopCloser(strings.NewReader(tc.input.content))
			a, err := NewFile(rc, tc.input.mode, tc.input.path, tc.input.state, tc.input.overwrite)
			require.NoError(t, err)
			err = inDirectory(workDir, func() error {
				return a.Apply(context.Background())
			})
			require.NoError(t, err)

			actualFiles, err := collectFilesInDirectory(workDir)
			require.NoError(t, err)
			assert.Equal(t, tc.wantFiles, actualFiles)
		})
	}
}

func ptr[T any](t T) *T {
	return &t
}
