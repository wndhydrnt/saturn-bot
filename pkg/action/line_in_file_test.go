package action

import (
	"errors"
	"testing"
)

var (
	testContent = "abc\ndef\n\nghi\njkl\n"
)

func TestLineDelete_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When state=delete and regex matches a line then it deletes the line from the file",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineDelete("test.txt", "^ghi$")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:  "When state=delete and regex matches multiple lines then it deletes each line from the file",
			files: map[string]string{"test.txt": "abc\nghi\ndef\n\nghi\njkl"},
			action: func() (Action, error) {
				return NewLineDelete("test.txt", "^ghi$")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:  "When state=delete and regex does not match a line then it leaves the file as-is",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineDelete("test.txt", "^xyz$")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:  "When state=delete and path contains a glob pattern then it deletes the line from each matching file",
			files: map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			action: func() (Action, error) {
				return NewLineDelete("*.txt", "^ghi$")
			},
			wantFiles: map[string]string{"test1.txt": "abc\ndef\n\njkl\n", "test2.txt": "abc\ndef\n\njkl\n"},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestLineInsert_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When state=insert then it adds the line at the end of the file",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("EOF", "ttt", "test.txt")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\nttt\n"},
		},
		{
			name:  "When state=insert and path is a glob pattern then it adds the line at the end of each file",
			files: map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("EOF", "ttt", "*.txt")
			},
			wantFiles: map[string]string{
				"test1.txt": "abc\ndef\n\nghi\njkl\nttt\n",
				"test2.txt": "abc\ndef\n\nghi\njkl\nttt\n",
			},
		},
		{
			name:  "When state=insert and insertAt=BOF then it adds the line at the end of the file",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("BOF", "ttt", "test.txt")
			},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:  "When state=insert and the value of insertAt is invalid then it errors",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("EXO", "unit test", "test.txt")
			},
			wantError: errors.New("invalid value EXO of parameter insertAt - can be one of BOF,EOF"),
		},
		{
			name:  "When state=insert and the value of insertAt isn't in upper-case then it still works",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("bOf", "ttt", "test.txt")
			},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:  "When state=insert and insertAt=BOF and the line is already present then there is no change",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("BOF", "abc", "test.txt")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:  "When state=insert and insertAt=EOF and the line is already present then there is no change",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineInsert("EOF", "jkl", "test.txt")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestLineReplace_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:  "When state=replace and regex matches a line then it replaces the line",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineReplace("ttt", "test.txt", "def")
			},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:  "When state=replace and path contains a glob pattern and regex matches a line then it replaces the line in each file",
			files: map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			action: func() (Action, error) {
				return NewLineReplace("ttt", "*.txt", "def")
			},
			wantFiles: map[string]string{"test1.txt": "abc\nttt\n\nghi\njkl\n", "test2.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:  "When state=replace and regex matches multiple lines then it replaces each line",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineReplace("ttt", "test.txt", "def|ghi")
			},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nttt\njkl\n"},
		},
		{
			name:  "When state=replace and regex contains match groups then it allows its usage",
			files: map[string]string{"test.txt": testContent},
			action: func() (Action, error) {
				return NewLineReplace("${1}T${2}", "test.txt", "(.+)h(.+)")
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\ngTi\njkl\n"},
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
