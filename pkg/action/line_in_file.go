package action

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wndhydrnt/saturn-bot/pkg/str"
)

const (
	beginningOfFile = "BOF"
	endOfFile       = "EOF"
)

type LineDeleteFactory struct{}

func (f LineDeleteFactory) Create(params map[string]string, _ string) (Action, error) {
	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	search := params["search"]
	regexpRaw := params["regexp"]
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
		err := forEachLine(path, func(line []byte) ([]byte, error) {
			if d.search != "" && string(line) == d.search {
				return nil, nil
			}

			if d.regex != nil && d.regex.Match(line) {
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

func (f LineInsertFactory) Create(params map[string]string, _ string) (Action, error) {
	insertAt := params["insertAt"]
	if insertAt == "" {
		insertAt = "EOF"
	}

	insertAt = strings.ToUpper(insertAt)
	if insertAt != beginningOfFile && insertAt != endOfFile {
		return nil, fmt.Errorf("invalid value of parameter `insertAt` - can be one of BOF,EOF")
	}

	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	line := params["line"]
	if line == "" {
		return nil, fmt.Errorf("required parameter `line` not set")
	}

	return &lineInsert{
		insertAt: insertAt,
		line:     line,
		path:     path,
	}, nil
}

func (f LineInsertFactory) Name() string {
	return "lineInsert"
}

type lineInsert struct {
	insertAt string
	line     string
	path     string
}

func (a *lineInsert) Apply(_ context.Context) error {
	paths, err := filepath.Glob(a.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern to insert line: %w", err)
	}

	for _, path := range paths {
		source, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file to insert line: %w", err)
		}

		defer source.Close()
		target, err := os.CreateTemp("", "*")
		if err != nil {
			return fmt.Errorf("create temporary file: %w", err)
		}

		defer target.Close()

		atBeginning := true
		var lastLine string
		scanner := bufio.NewScanner(source)
		for scanner.Scan() {
			if atBeginning {
				if a.insertAt == beginningOfFile && a.line != scanner.Text() {
					_, err := target.WriteString(a.line)
					if err != nil {
						return fmt.Errorf("write line at the beginning of temporary file: %w", err)
					}

					_, err = target.WriteString("\n")
					if err != nil {
						return fmt.Errorf("write linefeed at the beginning of temporary file: %w", err)
					}
				}

				atBeginning = false
			}

			_, err := target.Write(scanner.Bytes())
			if err != nil {
				return fmt.Errorf("write line to temporary file: %w", err)
			}

			_, err = target.WriteString("\n")
			if err != nil {
				return fmt.Errorf("write linefeed to temporary file: %w", err)
			}

			lastLine = scanner.Text()
		}

		if err = scanner.Err(); err != nil {
			return fmt.Errorf("scanner failed: %w", err)
		}

		if a.insertAt == endOfFile && a.line != lastLine {
			_, err := target.WriteString(a.line)
			if err != nil {
				return fmt.Errorf("write line at the end of temporary file: %w", err)
			}

			_, err = target.WriteString("\n")
			if err != nil {
				return fmt.Errorf("write linefeed at the end of temporary file: %w", err)
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

		_ = source.Close()
		_ = target.Close()
		err = os.Rename(target.Name(), source.Name())
		if err != nil {
			return fmt.Errorf("move target file to source file: %w", err)
		}
	}

	return nil
}

func (a *lineInsert) String() string {
	return fmt.Sprintf("lineInsert(insertAt=%s,line=%s,path=%s)", a.insertAt, a.line, a.path)
}

type LineReplaceFactory struct{}

func (f LineReplaceFactory) Create(params map[string]string, _ string) (Action, error) {
	path := params["path"]
	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	line := params["line"]
	if line == "" {
		return nil, fmt.Errorf("required parameter `line` not set")
	}

	regexpRaw := params["regexp"]
	search := params["search"]
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
		err := forEachLine(path, func(line []byte) ([]byte, error) {
			if a.search != "" && a.search == string(line) {
				return []byte(a.line), nil
			}

			if a.regexp != nil && a.regexp.Match(line) {
				new := a.regexp.ReplaceAll(line, []byte(a.line))
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

func forEachLine(filePath string, f func(line []byte) ([]byte, error)) error {
	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file to delete line: %w", err)
	}

	defer source.Close()
	target, err := os.CreateTemp("", "*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	defer target.Close()
	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		newB, err := f(scanner.Bytes())
		if err != nil {
			return err
		}

		if newB != nil {
			_, err = target.Write(newB)
			if err != nil {
				return fmt.Errorf("write line to temporary file: %w", err)
			}

			_, err = target.WriteString("\n")
			if err != nil {
				return fmt.Errorf("write linefeed to temporary file: %w", err)
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scanner failed: %w", err)
	}

	sourceStat, err := source.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	err = target.Chmod(sourceStat.Mode())
	if err != nil {
		return fmt.Errorf("chmod target file: %w", err)
	}

	_ = source.Close()
	_ = target.Close()
	err = os.Rename(target.Name(), source.Name())
	if err != nil {
		return fmt.Errorf("move target file to source file: %w", err)
	}

	return nil
}
