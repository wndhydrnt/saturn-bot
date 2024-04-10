package action

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
)

const (
	beginningOfFile = "BOF"
	endOfFile       = "EOF"
)

var (
	insertAtValues = []string{beginningOfFile, endOfFile}
)

func NewLineInFile(insertAt, line, path, regex, state string) (Action, error) {
	switch state {
	case "delete":
		regexC, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("compile regex: %w", err)
		}

		a := &deleteLineInFile{
			path:  path,
			regex: regexC,
		}
		return a, nil
	case "insert":
		if insertAt == "" {
			insertAt = endOfFile
		} else {
			insertAt = strings.ToUpper(insertAt)
		}

		if !slices.Contains(insertAtValues, insertAt) {
			return nil, fmt.Errorf("invalid value %s of parameter insertAt - can be one of %s", insertAt, strings.Join(insertAtValues, ","))
		}

		return &insertLineInFile{
			insertAt: insertAt,
			line:     line,
			path:     path,
		}, nil

	case "replace":
		regexC, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("compile regex: %w", err)
		}

		return &replaceLineInFile{
			line:  line,
			path:  path,
			regex: regexC,
		}, nil
	}

	return nil, fmt.Errorf("unknown value for state - can be one of delete,insert,replace was %s", state)
}

type deleteLineInFile struct {
	path  string
	regex *regexp.Regexp
}

func (d *deleteLineInFile) Apply(_ context.Context) error {
	err := forEachLine(d.path, func(line []byte) ([]byte, error) {
		if d.regex.Match(line) {
			return nil, nil
		}

		return line, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *deleteLineInFile) String() string {
	return fmt.Sprintf("lineInFile(path=%s,regex=%s,state=delete)", d.path, d.regex.String())
}

type insertLineInFile struct {
	insertAt string
	line     string
	path     string
}

func (a *insertLineInFile) Apply(_ context.Context) error {
	source, err := os.Open(a.path)
	if err != nil {
		return fmt.Errorf("open file to delete line: %w", err)
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

	return nil
}

func (a *insertLineInFile) String() string {
	return fmt.Sprintf("lineInFile(insertAt=%s,line=%s,path=%s,state=insert)", a.insertAt, a.line, a.path)
}

type replaceLineInFile struct {
	line  string
	path  string
	regex *regexp.Regexp
}

func (a *replaceLineInFile) Apply(_ context.Context) error {
	_ = forEachLine(a.path, func(line []byte) ([]byte, error) {
		if a.regex.Match(line) {
			new := a.regex.ReplaceAll(line, []byte(a.line))
			return new, nil
		}

		return line, nil
	})
	return nil
}

func (a *replaceLineInFile) String() string {
	return fmt.Sprintf("lineInFile(line=%s,path=%s,regex=%s,state=replace)", a.line, a.path, a.regex.String())
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
