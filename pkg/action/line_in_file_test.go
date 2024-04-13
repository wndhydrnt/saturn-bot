package action

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testContent = "abc\ndef\n\nghi\njkl\n"
)

func TestLineInFile_Apply(t *testing.T) {
	type input struct {
		insertAt string
		line     string
		path     string
		regex    string
		state    string
	}

	testCases := []struct {
		name      string
		files     map[string]string
		input     input
		wantError error
		wantFiles map[string]string
	}{
		{
			name:      "When state=delete and regex matches a line then it deletes the line from the file",
			files:     map[string]string{"test.txt": testContent},
			input:     input{path: "test.txt", regex: "^ghi$", state: "delete"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:      "When state=delete and regex matches multiple lines then it deletes each line from the file",
			files:     map[string]string{"test.txt": "abc\nghi\ndef\n\nghi\njkl"},
			input:     input{path: "test.txt", regex: "^ghi$", state: "delete"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:      "When state=delete and regex does not match a line then it leaves the file as-is",
			files:     map[string]string{"test.txt": testContent},
			input:     input{path: "test.txt", regex: "^xyz$", state: "delete"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:      "When state=delete and path contains a glob pattern then it deletes the line from each matching file",
			files:     map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			input:     input{path: "*.txt", regex: "^ghi$", state: "delete"},
			wantFiles: map[string]string{"test1.txt": "abc\ndef\n\njkl\n", "test2.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:      "When state=insert then it adds the line at the end of the file",
			files:     map[string]string{"test.txt": testContent},
			input:     input{line: "ttt", path: "test.txt", state: "insert"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\nttt\n"},
		},
		{
			name:  "When state=insert and path is a glob pattern then it adds the line at the end of each file",
			files: map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			input: input{line: "ttt", path: "*.txt", state: "insert"},
			wantFiles: map[string]string{
				"test1.txt": "abc\ndef\n\nghi\njkl\nttt\n",
				"test2.txt": "abc\ndef\n\nghi\njkl\nttt\n",
			},
		},
		{
			name:      "When state=insert and insertAt=BOF then it adds the line at the end of the file",
			files:     map[string]string{"test.txt": testContent},
			input:     input{insertAt: "BOF", line: "ttt", path: "test.txt", state: "insert"},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:      "When state=insert and the value of insertAt is invalid then it errors",
			files:     map[string]string{"test.txt": testContent},
			input:     input{insertAt: "EXO", line: "unit test", path: "test.txt", state: "insert"},
			wantError: errors.New("invalid value EXO of parameter insertAt - can be one of BOF,EOF"),
		},
		{
			name:      "When state=insert and the value of insertAt isn't in upper-case then it still works",
			files:     map[string]string{"test.txt": testContent},
			input:     input{insertAt: "bOf", line: "ttt", path: "test.txt", state: "insert"},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:      "When state=insert and insertAt=BOF and the line is already present then there is no change",
			files:     map[string]string{"test.txt": testContent},
			input:     input{insertAt: "BOF", line: "abc", path: "test.txt", state: "insert"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:      "When state=insert and insertAt=EOF and the line is already present then there is no change",
			files:     map[string]string{"test.txt": testContent},
			input:     input{insertAt: "EOF", line: "jkl", path: "test.txt", state: "insert"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:      "When state=replace and regex matches a line then it replaces the line",
			files:     map[string]string{"test.txt": testContent},
			input:     input{line: "ttt", path: "test.txt", regex: "def", state: "replace"},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:      "When state=replace and path contains a glob pattern and regex matches a line then it replaces the line in each file",
			files:     map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			input:     input{line: "ttt", path: "*.txt", regex: "def", state: "replace"},
			wantFiles: map[string]string{"test1.txt": "abc\nttt\n\nghi\njkl\n", "test2.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:      "When state=replace and regex matches multiple lines then it replaces each line",
			files:     map[string]string{"test.txt": testContent},
			input:     input{line: "ttt", path: "test.txt", regex: "def|ghi", state: "replace"},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nttt\njkl\n"},
		},
		{
			name:      "When state=replace and regex contains match groups then it allows its usage",
			files:     map[string]string{"test.txt": testContent},
			input:     input{line: "${1}T${2}", path: "test.txt", regex: "(.+)h(.+)", state: "replace"},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\ngTi\njkl\n"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workDir, err := setupTestFiles(tc.files)
			require.NoError(t, err)

			a, err := NewLineInFile(tc.input.insertAt, tc.input.line, tc.input.path, tc.input.regex, tc.input.state)
			if tc.wantError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantError.Error())
				return
			}

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
