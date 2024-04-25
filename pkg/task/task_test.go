package task

import (
	"fmt"
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
	err = tr.ReadAll([]string{f.Name()})
	require.NoError(t, err)

	assert.Len(t, tr.GetTasks(), 2)
	assert.Equal(t, "Task One", tr.GetTasks()[0].SourceTask().Name)
	assert.Equal(t, "Task Two", tr.GetTasks()[1].SourceTask().Name)
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
	err = tr.ReadAll([]string{f.Name()})

	assert.EqualError(t, err, fmt.Sprintf("failed to read tasks from file %s: unsupported extension: .yamll", f.Name()))
	assert.Len(t, tr.GetTasks(), 0)
}

func TestRegistry_ReadAll_AllSupportedActions(t *testing.T) {
	tasksRaw := `
name: Task
actions:
  fileCreate:
  - content: "Unit Test"
    path: unit-test.txt
  - content: "$file: content.txt"
    path: unit-test-content.txt
  fileDelete:
  - path: delete.txt
  lineDelete:
  - path: example.txt
    line: "to delete"
  lineInsert:
  - path: example.txt
    line: "Unit Test"
  lineReplace:
  - path: example.txt
    line: "to replace"
    search: "to find"
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
	err = tr.ReadAll([]string{taskPath})
	require.NoError(t, err)

	require.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, "Task", task.SourceTask().Name)
	wantActions := []string{
		"fileCreate(mode=644,overwrite=true,path=unit-test.txt)",
		"fileCreate(mode=644,overwrite=true,path=unit-test-content.txt)",
		"fileDelete(path=delete.txt)",
		"lineDelete(line=example.txt,path=to delete)",
		"lineInsert(insertAt=EOF,line=Unit Test,path=example.txt)",
		"lineReplace(line=to replace,path=example.txt,search=to find)",
	}
	var actualActions []string
	for _, a := range task.Actions() {
		actualActions = append(actualActions, a.String())
	}

	assert.Equal(t, wantActions, actualActions)
}

func TestRegistry_ReadAll_AllSupportedFilters(t *testing.T) {
	tasksRaw := `
  name: Task
  filters:
    repositoryName:
      - names: ["git.localhost/unit/test", "git.localhost/unit/test2"]
    file:
      - path: unit-test.txt
    fileContent:
      - path: hello-world.txt
        search: Hello World
`
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	taskPath := filepath.Join(tempDir, "task.yaml")
	taskFile, err := os.Create(taskPath)
	require.NoError(t, err)
	_, err = taskFile.WriteString(tasksRaw)
	require.NoError(t, err)
	taskFile.Close()

	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	tr := &Registry{}
	err = tr.ReadAll([]string{taskPath})
	require.NoError(t, err)

	require.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, "Task", task.SourceTask().Name)
	wantFilters := []string{
		"repositoryName(names=[^git.localhost/unit/test$,^git.localhost/unit/test2$])",
		"file(path=unit-test.txt)",
		"fileContent(path=hello-world.txt,search=Hello World)",
	}
	var actualFilters []string
	for _, a := range task.Filters() {
		actualFilters = append(actualFilters, a.String())
	}

	assert.Equal(t, wantFilters, actualFilters)
}
