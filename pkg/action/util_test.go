package action

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name           string
	bootstrap      func(ctx context.Context) (string, context.Context)
	files          map[string]string
	factory        Factory
	params         map[string]any
	wantError      error
	wantErrorApply error
	wantFiles      map[string]string
}

func collectFilesInDirectory(dir string) (map[string]string, error) {
	fileMap := map[string]string{}
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		withoutBase := strings.TrimPrefix(path, dir+"/")
		fileMap[withoutBase] = string(b)
		return nil
	})
	return fileMap, err
}

func inDirectory(dir string, f func() error) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("change to work directory: %w", err)
	}

	funcErr := f()
	err = os.Chdir(currentDir)
	if err != nil {
		return fmt.Errorf("changing back to previous directory: %w", err)
	}

	return funcErr
}

func runTestCase(t *testing.T, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		workDir, err := setupTestFiles(tc.files)
		require.NoError(t, err)

		ctx := context.Background()
		taskPath := ""
		if tc.bootstrap != nil {
			taskPath, ctx = tc.bootstrap(ctx)
		}

		a, err := tc.factory.Create(tc.params, taskPath)
		if tc.wantError == nil {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.wantError.Error())
			return
		}

		err = inDirectory(workDir, func() error {
			return a.Apply(ctx)
		})
		if tc.wantErrorApply == nil {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tc.wantErrorApply.Error())
			return
		}

		actualFiles, err := collectFilesInDirectory(workDir)
		require.NoError(t, err)
		assert.Equal(t, tc.wantFiles, actualFiles)
	})
}

func setupTestFiles(fileMap map[string]string) (string, error) {
	workDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	for name, content := range fileMap {
		path := filepath.Join(workDir, name)
		f, err := os.Create(path)
		if err != nil {
			return "", err
		}

		_, err = f.WriteString(content)
		if err != nil {
			return "", err
		}
	}

	return workDir, nil
}
