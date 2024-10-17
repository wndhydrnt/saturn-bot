package task_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

func TestRegistry_ReadAll(t *testing.T) {
	tasksRaw := `
name: Task One
---
name: Task Two
`

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	f, err := os.CreateTemp(tempDir, "*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(tasksRaw)
	require.NoError(t, err)
	f.Close()
	defer func() {
		err := os.Remove(f.Name())
		require.NoError(t, err)
	}()

	tr := &task.Registry{}
	pathGlob := fmt.Sprintf("%s/*.yaml", filepath.Dir(f.Name()))
	err = tr.ReadAll([]string{pathGlob})
	require.NoError(t, err)

	assert.Len(t, tr.GetTasks(), 2)
	assert.Equal(t, "Task One", tr.GetTasks()[0].Name)
	assert.Equal(t, "Task Two", tr.GetTasks()[1].Name)
}

func TestRegistry_ReadAll_AllBuiltInActions(t *testing.T) {
	tasksRaw := `
name: Task
actions:
  - action: fileCreate
    params:
      content: "Unit Test"
      path: unit-test.txt
  - action: fileCreate
    params:
      contentFromFile: content.txt
      path: unit-test-content.txt
  - action: fileDelete
    params:
      path: delete.txt
  - action: lineDelete
    params:
      search: "to delete"
      path: example.txt
  - action: lineInsert
    params:
      line: "Unit Test"
      path: example.txt
  - action: lineReplace
    params:
      path: example.txt
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

	tr := task.NewRegistry(options.Opts{ActionFactories: action.BuiltInFactories})
	err = tr.ReadAll([]string{taskPath})
	require.NoError(t, err)

	require.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, "Task", task.Name)
	wantActions := []string{
		"fileCreate(mode=644,overwrite=true,path=unit-test.txt)",
		"fileCreate(mode=644,overwrite=true,path=unit-test-content.txt)",
		"fileDelete(path=delete.txt)",
		"lineDelete(path=example.txt,search=to delete)",
		"lineInsert(insertAt=EOF,line=Unit Test,path=example.txt)",
		"lineReplace(line=to replace,path=example.txt,search=to find)",
	}
	var actualActions []string
	for _, a := range task.Actions() {
		actualActions = append(actualActions, a.String())
	}

	assert.Equal(t, wantActions, actualActions)
}

func TestRegistry_ReadAll_AllBuiltInFilters(t *testing.T) {
	tasksRaw := `
  name: Task
  filters:
    - filter: repository
      params:
        host: git.localhost
        owner: unit
        name: test|test2
    - filter: file
      params:
        paths: [unit-test.txt]
    - filter: fileContent
      params:
        path: hello-world.txt
        regexp: Hello World
    - filter: file
      params:
        paths: [test.txt]
      reverse: true
    - filter: jsonpath
      params:
        expressions: ["$.dependencies"]
        path: package.json
    - filter: xpath
      params:
        expressions: ["//project"]
        path: pom.xml
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

	tr := task.NewRegistry(options.Opts{FilterFactories: filter.BuiltInFactories})
	err = tr.ReadAll([]string{taskPath})
	require.NoError(t, err)

	require.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, "Task", task.Name)
	wantFilters := []string{
		"repository(host=^git.localhost$,owner=^unit$,name=^test|test2$)",
		"file(op=and,paths=[unit-test.txt])",
		"fileContent(path=hello-world.txt,regexp=Hello World)",
		"!file(op=and,paths=[test.txt])",
		"jsonpath(expressions=[$.dependencies],path=package.json)",
		"xpath(expressions=[//project],path=pom.xml)",
	}
	var actualFilters []string
	for _, a := range task.Filters() {
		actualFilters = append(actualFilters, a.String())
	}

	assert.Equal(t, wantFilters, actualFilters)
}

func TestRegistry_ReadAll_DeactivatedTask(t *testing.T) {
	tasksRaw := `
name: Task One
active: false
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

	tr := &task.Registry{}
	err = tr.ReadAll([]string{f.Name()})
	require.NoError(t, err)

	assert.Len(t, tr.GetTasks(), 1)
	assert.Equal(t, "Task Two", tr.GetTasks()[0].Name)
}

func TestRegistry_ReadAll_InvalidSchedule(t *testing.T) {
	tasksRaw := `
name: Task Schedule Invalid
schedule: "* * * *"
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

	tr := &task.Registry{}
	err = tr.ReadAll([]string{f.Name()})
	require.Errorf(t, err, "failed to read tasks from file %s: parse schedule: Expected exactly 5 fields, found 4: * * * *", f.Name())
}

func TestRegistry_ReadAll_SortRepositoryFilterFirst(t *testing.T) {
	tasksRaw := `
  name: Task
  filters:
    - filter: file
      params:
        paths: [unit-test.txt]
    - filter: repository
      params:
        host: git.localhost
        owner: unit
        name: test
    - filter: repository
      params:
        host: git.localhost
        owner: unit
        name: test2
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

	tr := task.NewRegistry(options.Opts{FilterFactories: filter.BuiltInFactories})
	err = tr.ReadAll([]string{f.Name()})
	require.NoError(t, err)

	assert.Len(t, tr.GetTasks(), 1)
	task := tr.GetTasks()[0]
	assert.Equal(t, task.Filters()[0].String(), "repository(host=^git.localhost$,owner=^unit$,name=^test$)")
	assert.Equal(t, task.Filters()[1].String(), "repository(host=^git.localhost$,owner=^unit$,name=^test2$)")
	assert.Equal(t, task.Filters()[2].String(), "file(op=and,paths=[unit-test.txt])")
}
