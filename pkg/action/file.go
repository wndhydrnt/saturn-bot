package action

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func (f FileCreateFactory) Create(params map[string]any, taskPath string) (Action, error) {
	var content string
	if params["content"] != nil {
		contentCast, ok := params["content"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `content` is of type %T not string", params["content"])
		}

		content = contentCast
	}

	var contentFromFile string
	if params["contentFromFile"] != nil {
		contentFromFileCast, ok := params["contentFromFile"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `contentFromFile` is of type %T not string", params["contentFromFile"])
		}

		contentFromFile = contentFromFileCast
	}

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

	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["path"])
	}

	var mode fs.FileMode
	if params["mode"] == nil {
		mode = 0644
	} else {
		modeInt, ok := params["mode"].(int)
		if !ok {
			return nil, fmt.Errorf("parameter `mode` is of type %T not int", params["mode"])
		}

		mode = fs.FileMode(modeInt)
	}

	var overwrite bool
	if params["overwrite"] == nil {
		overwrite = true
	} else {
		overwriteRaw, ok := params["overwrite"].(bool)
		if !ok {
			return nil, fmt.Errorf("parameter `overwrite` is of type %T not bool", params["overwrite"])
		}
		overwrite = overwriteRaw
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

		defer file.Close()
		_, err = io.Copy(file, a.content)
		if err != nil {
			return fmt.Errorf("write content to file %s: %w", a.path, err)
		}

		err = file.Chmod(a.mode)
		if err != nil {
			return fmt.Errorf("change mode of file: %w", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("close file %s: %w", a.path, err)
		}
	}

	return nil
}

func (a *fileCreate) String() string {
	return fmt.Sprintf("fileCreate(mode=%o,overwrite=%t,path=%s)", a.mode, a.overwrite, a.path)
}

type FileDeleteFactory struct{}

func (f FileDeleteFactory) Create(params map[string]any, _ string) (Action, error) {
	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["path"])
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
