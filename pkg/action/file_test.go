package action

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-sync/pkg/task/schema"
)

func TestFileCreate_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When field content is set and overwrite=false and the file does not exist then it creates the file",
			files: map[string]string{},
			action: func() (Action, error) {
				content := "abc\n"
				return NewFileCreateFromTask(schema.TaskActionsFileCreateElem{
					Content: &content,
					Path:    "test.txt",
				}, "")
			},
			wantError: nil,
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When field contentFile is set and overwrite=false and the file does not exist then it creates the file",
			files: map[string]string{},
			action: func() (Action, error) {
				tmpDir, err := os.MkdirTemp("", "*")
				require.NoError(t, err)
				contentFilePath := "content.txt"
				f, err := os.Create(filepath.Join(tmpDir, contentFilePath))
				require.NoError(t, err)
				_, err = f.WriteString("abc\n")
				require.NoError(t, err)
				_ = f.Close()
				return NewFileCreateFromTask(schema.TaskActionsFileCreateElem{
					ContentFile: &contentFilePath,
					Path:        "test.txt",
				}, filepath.Join(tmpDir, "task.yaml"))
			},
			wantError: nil,
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When the file exists and overwrite=false then it does not update the file",
			files: map[string]string{"test.txt": "abc\n"},
			action: func() (Action, error) {
				content := "def\n"
				return NewFileCreateFromTask(schema.TaskActionsFileCreateElem{
					Content:   &content,
					Overwrite: false,
					Path:      "test.txt",
				}, "")
			},
			wantFiles: map[string]string{"test.txt": "abc\n"},
		},
		{
			name:  "When state=present and overwrite=true and the file exists then it updates the file",
			files: map[string]string{"test.txt": "abc\n"},
			action: func() (Action, error) {
				content := "def\n"
				return NewFileCreateFromTask(schema.TaskActionsFileCreateElem{
					Content:   &content,
					Overwrite: true,
					Path:      "test.txt",
				}, "")
			},
			wantFiles: map[string]string{"test.txt": "def\n"},
		},
		{
			name:  "When overwrite=true and the file does not exist then it creates the file",
			files: map[string]string{},
			action: func() (Action, error) {
				content := "def\n"
				return NewFileCreateFromTask(schema.TaskActionsFileCreateElem{
					Content:   &content,
					Overwrite: true,
					Path:      "test.txt",
				}, "")
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
