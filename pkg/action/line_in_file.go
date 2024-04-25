package action

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	beginningOfFile = "BOF"
	endOfFile       = "EOF"
)

func NewLineDelete(path, line string) (Action, error) {
	regexC, err := regexp.Compile(line)
	if err != nil {
		return nil, fmt.Errorf("compile regex: %w", err)
	}

	return &lineDelete{
		path:  path,
		regex: regexC,
	}, nil
}

type lineDelete struct {
	path  string
	regex *regexp.Regexp
}

func (d *lineDelete) Apply(_ context.Context) error {
	paths, err := filepath.Glob(d.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern: %w", err)
	}

	for _, path := range paths {
		err := forEachLine(path, func(line []byte) ([]byte, error) {
			if d.regex.Match(line) {
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
	return fmt.Sprintf("lineDelete(line=%s,path=%s)", d.regex.String(), d.path)
}

func NewLineInsert(insertAt, line, path string) (Action, error) {
	insertAt = strings.ToUpper(insertAt)
	if insertAt != beginningOfFile && insertAt != endOfFile {
		return nil, fmt.Errorf("invalid value %s of parameter insertAt - can be one of BOF,EOF", insertAt)
	}

	return &lineInsert{insertAt: insertAt, line: line, path: path}, nil
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

func NewLineReplace(line, path, search string) (Action, error) {
	regexC, err := regexp.Compile(search)
	if err != nil {
		return nil, fmt.Errorf("compile regex: %w", err)
	}

	return &lineReplace{
		line:  line,
		path:  path,
		regex: regexC,
	}, nil
}

type lineReplace struct {
	line  string
	path  string
	regex *regexp.Regexp
}

func (a *lineReplace) Apply(_ context.Context) error {
	paths, err := filepath.Glob(a.path)
	if err != nil {
		return fmt.Errorf("parse glob pattern to replace line: %w", err)
	}

	for _, path := range paths {
		err := forEachLine(path, func(line []byte) ([]byte, error) {
			if a.regex.Match(line) {
				new := a.regex.ReplaceAll(line, []byte(a.line))
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
	return fmt.Sprintf("lineReplace(line=%s,path=%s,search=%s)", a.line, a.path, a.regex.String())
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
