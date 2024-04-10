package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_ReadAll(t *testing.T) {
	tasksRaw := `
name: Task One
---
name: Task Two
`

	f, err := os.CreateTemp("", "*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(tasksRaw)
	require.NoError(t, err)
	f.Close()
	defer func() {
		err := os.Remove(f.Name())
		require.NoError(t, err)
	}()

	tr := &Registry{}
	tr.ReadAll([]string{f.Name()})

	assert.Len(t, tr.GetTasks(), 2)
	assert.Equal(t, "Task One", tr.GetTasks()[0].SourceTask().GetName())
	assert.Equal(t, "Task Two", tr.GetTasks()[1].SourceTask().GetName())
}

func TestRegistry_ReadAll_Unsupported(t *testing.T) {
	tasksRaw := `
name: Task One
---
name: Task Two
`

	f, err := os.CreateTemp("", "*.yamll")
	require.NoError(t, err)
	_, err = f.WriteString(tasksRaw)
	require.NoError(t, err)
	f.Close()
	defer func() {
		err := os.Remove(f.Name())
		require.NoError(t, err)
	}()

	tr := &Registry{}
	tr.ReadAll([]string{f.Name()})

	assert.Len(t, tr.GetTasks(), 0)
}

func TestRegistry_ReadAll_AllSupportedActions(t *testing.T) {
	tasksRaw := `
name: Task
actions:
  file:
  - state: present
    content: "Unit Test"
    path: unit-test.txt
  - state: present
    content: "$file: content.txt"
    path: unit-test-content.txt
  lineInFile:
  - path: example.txt
    state: insert
    line: "Unit Test"
`

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	taskPath := filepath.Join(tempDir, "task.yaml")
	taskFile, err := os.Create(taskPath)
	require.NoError(t, err)
	_, err = taskFile.WriteString(tasksRaw)
	require.NoError(t, err)
	taskFile.Close()

	contentPath := filepath.Join(tempDir, "content.txt")
	contentFile, err := os.Create(contentPath)
	require.NoError(t, err)
	_, err = contentFile.WriteString("Unit Test Content File")
	require.NoError(t, err)
	contentFile.Close()

	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	tr := &Registry{}
	tr.ReadAll([]string{taskPath})

	require.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, "Task", task.SourceTask().GetName())
	wantActions := []string{
		"file(overwrite=true,path=unit-test.txt,state=present)",
		"file(overwrite=true,path=unit-test-content.txt,state=present)",
		"lineInFile(insertAt=EOF,line=Unit Test,path=example.txt,state=insert)",
	}
	var actualActions []string
	for _, a := range task.Actions() {
		actualActions = append(actualActions, a.String())
	}

	assert.Equal(t, wantActions, actualActions)
}
