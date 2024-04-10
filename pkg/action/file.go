package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

func NewFile(content io.ReadCloser, mode, path, state string, overwrite *bool) (Action, error) {
	if path == "" {
		return nil, fmt.Errorf("property 'path' of action 'file' cannot be empty")
	}

	switch state {
	case "absent":
		return &deleteFile{path: path}, nil
	case "present":
		overwriteReal := true
		if overwrite != nil {
			overwriteReal = *overwrite
		}

		return &createFile{
			content:   content,
			overwrite: overwriteReal,
			path:      path,
		}, nil
	}

	return nil, fmt.Errorf("unknown value for state - can be one of present,absent was %s", state)
}

type createFile struct {
	content   io.ReadCloser
	overwrite bool
	path      string
}

func (a *createFile) Apply(_ context.Context) error {
	defer a.content.Close()
	_, err := os.Stat(a.path)
	fileExists := true
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fileExists = false
		} else {
			return fmt.Errorf("check if file exists: %w", err)
		}
	}

	if a.overwrite || !fileExists {
		file, err := os.Create(a.path)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, a.content)
		if err != nil {
			return fmt.Errorf("write content to file %s: %w", a.path, err)
		}

		err = file.Sync()
		if err != nil {
			return fmt.Errorf("sync content to file %s: %w", a.path, err)
		}
	}

	return nil
}

func (a *createFile) String() string {
	return fmt.Sprintf("file(overwrite=%t,path=%s,state=present)", a.overwrite, a.path)
}

type deleteFile struct {
	path string
}

func (a *deleteFile) Apply(_ context.Context) error {
	err := os.Remove(a.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File already deleted. That is okay.
			return nil
		}

		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func (a *deleteFile) String() string {
	return fmt.Sprintf("file(path=%s,state=absent)", a.path)
}
