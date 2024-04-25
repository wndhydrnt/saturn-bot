package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

func NewFileCreate(content io.ReadCloser, mode int, overwrite bool, path string) Action {
	return &fileCreate{
		content:   content,
		mode:      mode,
		overwrite: overwrite,
		path:      path,
	}
}

type fileCreate struct {
	content   io.ReadCloser
	mode      int
	overwrite bool
	path      string
}

func (a *fileCreate) Apply(_ context.Context) error {
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

func (a *fileCreate) String() string {
	return fmt.Sprintf("fileCreate(mode=%d,overwrite=%t,path=%s)", a.mode, a.overwrite, a.path)
}

func NewFileDelete(path string) Action {
	return &fileDelete{path: path}
}

type fileDelete struct {
	path string
}

func (a *fileDelete) Apply(_ context.Context) error {
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

func (a *fileDelete) String() string {
	return fmt.Sprintf("fileDelete(path=%s)", a.path)
}
