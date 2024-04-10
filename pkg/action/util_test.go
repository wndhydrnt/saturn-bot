package action

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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
