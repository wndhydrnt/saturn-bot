package action

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"github.com/wndhydrnt/saturn-bot/pkg/str"
)

const (
	beginningOfFile = "BOF"
	endOfFile       = "EOF"
)

var (
	lineFeed = []byte("\n")
)

type LineDeleteFactory struct{}

func (f LineDeleteFactory) Create(params params.Params, _ string) (Action, error) {
	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["path"])
	}

	var search string
	if params["search"] != nil {
		searchCast, ok := params["search"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `search` is of type %T not string", params["search"])
		}
		search = searchCast
	}

	var regexpRaw string
	if params["regexp"] != nil {
		regexpCast, ok := params["regexp"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `regexp` is of type %T not string", params["regexp"])
		}
		regexpRaw = regexpCast
	}

	if search == "" && regexpRaw == "" {
		return nil, fmt.Errorf("either parameter `regexp` or `search` must be set")
	}

	if search != "" && regexpRaw != "" {
		return nil, fmt.Errorf("parameters `regexp` and `search` cannot bet set both")
	}

	var re *regexp.Regexp
	if regexpRaw != "" {
		var err error
		re, err = regexp.Compile(str.EncloseRegex(regexpRaw))
		if err != nil {
			return nil, fmt.Errorf("compile parameter `regexp` to regular expression: %w", err)
		}
	}

	return &lineDelete{
		path:   path,
		regex:  re,
		search: search,
	}, nil
}

func (f LineDeleteFactory) Name() string {
	return "lineDelete"
}

type lineDelete struct {
	path   string
	regex  *regexp.Regexp
	search string
}

func (d *lineDelete) Apply(_ context.Context) error {
	paths, err := filepath.Glob(d.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern: %w", err)
	}

	for _, path := range paths {
		err := forEachLine(path, func(line []byte, _ bool) ([]byte, error) {
			lineTrim := bytes.TrimSuffix(line, lineFeed)
			if d.search != "" && string(lineTrim) == d.search {
				return nil, nil
			}

			if d.regex != nil && d.regex.Match(lineTrim) {
				return nil, nil
			}

			return line, nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *lineDelete) String() string {
	if d.search != "" {
		return fmt.Sprintf("lineDelete(path=%s,search=%s)", d.path, d.search)
	}

	return fmt.Sprintf("lineDelete(path=%s,regexp=%s)", d.path, d.regex.String())
}

type LineInsertFactory struct{}

func (f LineInsertFactory) Create(params params.Params, _ string) (Action, error) {
	insertAt := "EOF"
	if params["insertAt"] != nil {
		insertAtCast, ok := params["insertAt"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `insertAt` is of type %T not string", params["insertAt"])
		}

		insertAt = insertAtCast
	}

	insertAt = strings.ToUpper(insertAt)
	if insertAt != beginningOfFile && insertAt != endOfFile {
		return nil, fmt.Errorf("invalid value of parameter `insertAt` - can be one of BOF,EOF")
	}

	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["path"])
	}

	if params["line"] == nil {
		return nil, fmt.Errorf("required parameter `line` not set")
	}
	line, ok := params["line"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["line"])
	}

	return &lineInsert{
		insertAt: insertAt,
		line:     []byte(line),
		path:     path,
	}, nil
}

func (f LineInsertFactory) Name() string {
	return "lineInsert"
}

type lineInsert struct {
	insertAt string
	line     []byte
	path     string
}

func (a *lineInsert) Apply(_ context.Context) error {
	paths, err := filepath.Glob(a.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern to insert line: %w", err)
	}

	for _, path := range paths {
		atBeginning := true
		err := forEachLine(path, func(line []byte, eof bool) ([]byte, error) {
			lineTrim := bytes.TrimSuffix(line, lineFeed)
			if atBeginning {
				atBeginning = false
				if a.insertAt == beginningOfFile && !bytes.Equal(lineTrim, []byte(a.line)) {
					result := []byte{}
					result = append(result, a.line...)
					result = append(result, lineFeed...)
					result = append(result, line...)
					return result, nil
				}
			}

			if a.insertAt == endOfFile && eof {
				result := []byte{}
				if len(line) == 0 {
					// Length of 0 indicates that the last line of the file ends in a newline.
					// Add the line and a newline in that case.
					result = append(result, a.line...)
					result = append(result, lineFeed...)
					return result, nil
				} else {
					// Do not add a newline at the end of the file
					// if the original file doesn't do so.
					result = append(result, line...)
					result = append(result, lineFeed...)
					result = append(result, a.line...)
					return result, nil
				}
			}

			return line, nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (a *lineInsert) String() string {
	return fmt.Sprintf("lineInsert(insertAt=%s,line=%s,path=%s)", a.insertAt, a.line, a.path)
}

type LineReplaceFactory struct{}

func (f LineReplaceFactory) Create(params params.Params, _ string) (Action, error) {
	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["path"])
	}

	if params["line"] == nil {
		return nil, fmt.Errorf("required parameter `line` not set")
	}
	line, ok := params["line"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `path` is of type %T not string", params["line"])
	}

	var search string
	if params["search"] != nil {
		searchCast, ok := params["search"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `search` is of type %T not string", params["search"])
		}
		search = searchCast
	}

	var regexpRaw string
	if params["regexp"] != nil {
		regexpCast, ok := params["regexp"].(string)
		if !ok {
			return nil, fmt.Errorf("parameter `regexp` is of type %T not string", params["regexp"])
		}
		regexpRaw = regexpCast
	}

	if search == "" && regexpRaw == "" {
		return nil, fmt.Errorf("either parameter `regexp` or `search` must be set")
	}

	if search != "" && regexpRaw != "" {
		return nil, fmt.Errorf("parameters `regexp` and `search` cannot be set both")
	}

	var re *regexp.Regexp
	if regexpRaw != "" {
		var err error
		re, err = regexp.Compile(str.EncloseRegex(regexpRaw))
		if err != nil {
			return nil, fmt.Errorf("compile parameter `regexp` to regular expression: %w", err)
		}
	}

	return &lineReplace{
		line:   line,
		path:   path,
		regexp: re,
		search: search,
	}, nil
}

func (f LineReplaceFactory) Name() string {
	return "lineReplace"
}

type lineReplace struct {
	line   string
	path   string
	regexp *regexp.Regexp
	search string
}

func (a *lineReplace) Apply(_ context.Context) error {
	paths, err := filepath.Glob(a.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern to replace line: %w", err)
	}

	for _, path := range paths {
		err := forEachLine(path, func(line []byte, _ bool) ([]byte, error) {
			lineTrim := bytes.TrimSuffix(line, lineFeed)
			if a.search != "" && a.search == string(lineTrim) {
				new := []byte(a.line)
				if bytes.HasSuffix(line, lineFeed) {
					new = append(new, lineFeed...)
				}

				return new, nil
			}

			if a.regexp != nil && a.regexp.Match(lineTrim) {
				new := a.regexp.ReplaceAll(lineTrim, []byte(a.line))
				if bytes.HasSuffix(line, lineFeed) {
					new = append(new, lineFeed...)
				}

				return new, nil
			}

			return line, nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *lineReplace) String() string {
	if a.search != "" {
		return fmt.Sprintf("lineReplace(line=%s,path=%s,search=%s)", a.line, a.path, a.search)
	}

	return fmt.Sprintf("lineReplace(line=%s,path=%s,regexp=%s)", a.line, a.path, a.regexp.String())
}

func forEachLine(filePath string, f func(line []byte, eof bool) ([]byte, error)) error {
	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file to delete line: %w", err)
	}

	defer source.Close()
	target, err := os.CreateTemp(filepath.Dir(filePath), "*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	defer target.Close()

	isEOF := false
	reader := bufio.NewReader(source)
	for {
		oldB, readErr := reader.ReadBytes('\n')
		if errors.Is(readErr, io.EOF) {
			isEOF = true
		}

		newB, err := f(oldB, isEOF)
		if err != nil {
			return err
		}

		if newB != nil {
			_, err = target.Write(newB)
			if err != nil {
				return fmt.Errorf("write line to temporary file: %w", err)
			}
		}

		if readErr != nil {
			break
		}
	}

	sourceStat, err := source.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	err = target.Chmod(sourceStat.Mode())
	if err != nil {
		return fmt.Errorf("chmod target file: %w", err)
	}

	err = source.Close()
	if err != nil {
		return fmt.Errorf("close source file: %w", err)
	}

	err = target.Close()
	if err != nil {
		return fmt.Errorf("close target file: %w", err)
	}

	err = os.Rename(target.Name(), source.Name())
	if err != nil {
		return fmt.Errorf("move target file to source file: %w", err)
	}

	return nil
}
