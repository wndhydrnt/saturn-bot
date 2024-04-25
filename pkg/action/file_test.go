package action

import (
	"io"
	"strings"
	"testing"
)

func TestFileCreate_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When state=present and overwrite=false and the file does not exist then it creates the file",
			files: map[string]string{},
			action: func() (Action, error) {
				content := io.NopCloser(strings.NewReader("abc\n"))
				return NewFileCreate(content, 644, false, "test.txt"), nil
			},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When state=present and overwrite=false and the file exists then it does not update the file",
			files: map[string]string{"test.txt": "abc\n"},
			action: func() (Action, error) {
				content := io.NopCloser(strings.NewReader("def\n"))
				return NewFileCreate(content, 644, false, "test.txt"), nil
			},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When state=present and overwrite=true and the file exists then it updates the file",
			files: map[string]string{"test.txt": "abc\n"},
			action: func() (Action, error) {
				content := io.NopCloser(strings.NewReader("def\n"))
				return NewFileCreate(content, 644, true, "test.txt"), nil
			},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:  "When state=present and overwrite=true and the file does not exist then it creates the file",
			files: map[string]string{},
			action: func() (Action, error) {
				content := io.NopCloser(strings.NewReader("def\n"))
				return NewFileCreate(content, 644, true, "test.txt"), nil
			},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestFileDelete_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When state=absent the file exists then it deletes the file",
			files: map[string]string{"test.txt": "abc\n"},
			action: func() (Action, error) {
				return NewFileDelete("test.txt"), nil
			},
			wantFiles: map[string]string{},
		},
		{
			name:  "When state=absent the file does not exist then it does nothing",
			files: map[string]string{},
			action: func() (Action, error) {
				return NewFileDelete("test.txt"), nil
			},
			wantFiles: map[string]string{},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
