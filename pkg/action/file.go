package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func createContentReader(taskPath, value string) (io.ReadCloser, error) {
	filePath := strings.TrimSpace(strings.TrimPrefix(value, "$file:"))
	taskDir := filepath.Dir(taskPath)
	abs, err := filepath.Abs(filepath.Join(taskDir, filePath))
	if err != nil {
		return nil, fmt.Errorf("get absolute path of %s: %w", filePath, err)
	}

	if !strings.HasPrefix(abs, taskDir) {
		return nil, fmt.Errorf("path escapes directory of task")
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open content reader: %w", err)
	}

	return f, nil
}

type FileCreateFactory struct{}

func (f FileCreateFactory) Create(params map[string]string, taskPath string) (Action, error) {
	content := params["content"]
	contentFromFile := params["contentFromFile"]

	if content == "" && contentFromFile == "" {
		return nil, fmt.Errorf("either parameter `content` or `contentFromFile` is required")
	}

	if content != "" && contentFromFile != "" {
		return nil, fmt.Errorf("parameters `content` and `contentFromFile` cannot be set at the same time")
	}

	var contentReader io.ReadCloser
	if content != "" {
		contentReader = io.NopCloser(strings.NewReader(content))
	} else {
		var err error
		contentReader, err = createContentReader(taskPath, contentFromFile)
		if err != nil {
			return nil, fmt.Errorf("read source of `contentFromFile`: %w", err)
		}
	}

	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	modeRaw := params["mode"]
	var mode fs.FileMode
	if modeRaw == "" {
		mode = 0644
	} else {
		modeConv, err := strconv.ParseUint(modeRaw, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("parse value of parameter `mode`: %w", err)
		}

		mode = fs.FileMode(modeConv)
	}

	overwriteRaw := params["overwrite"]
	overwrite := true
	if overwriteRaw == "false" {
		overwrite = false
	}

	return &fileCreate{
		content:   contentReader,
		mode:      mode,
		overwrite: overwrite,
		path:      path,
	}, nil
}

func (f FileCreateFactory) Name() string {
	return "fileCreate"
}

type fileCreate struct {
	content   io.ReadCloser
	mode      fs.FileMode
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

	if !fileExists {
		d := filepath.Dir(a.path)
		err := os.MkdirAll(d, 0755)
		if err != nil {
			return fmt.Errorf("create directory because it does not exist: %w", err)
		}
	}

	if a.overwrite || !fileExists {
		file, err := os.Create(a.path)
		if err != nil {
			return fmt.Errorf("create or truncate file: %w", err)
		}

		_, err = io.Copy(file, a.content)
		if err != nil {
			return fmt.Errorf("write content to file %s: %w", a.path, err)
		}

		err = file.Sync()
		if err != nil {
			return fmt.Errorf("sync content to file %s: %w", a.path, err)
		}

		err = file.Chmod(a.mode)
		if err != nil {
			return fmt.Errorf("change mode of file: %w", err)
		}
	}

	return nil
}

func (a *fileCreate) String() string {
	return fmt.Sprintf("fileCreate(mode=%d,overwrite=%t,path=%s)", a.mode, a.overwrite, a.path)
}

type FileDeleteFactory struct{}

func (f FileDeleteFactory) Create(params map[string]string, _ string) (Action, error) {
	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return &fileDelete{
		path: path,
	}, nil
}

func (f FileDeleteFactory) Name() string {
	return "fileDelete"
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
