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
			name:    "When regexp matches a line then it deletes the line",
			files:   map[string]string{"test.txt": testContent},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"regexp": "^ghi$",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:    "When regexp matches multiple lines then it deletes each line",
			files:   map[string]string{"test.txt": "abc\nghi\ndef\n\nghi\njkl"},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"regexp": "^ghi$",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:    "When regexp does not match a line then it leaves the file as-is",
			files:   map[string]string{"test.txt": testContent},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"regexp": "^xyz$",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:    "When parameter `path` contains a glob pattern then it deletes the line from each matching file",
			files:   map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "*.txt",
				"regexp": "^ghi$",
			},
			wantFiles: map[string]string{"test1.txt": "abc\ndef\n\njkl\n", "test2.txt": "abc\ndef\n\njkl\n"},
		},
		{
			name:    "When parameter `path` is not set then it errors",
			files:   map[string]string{},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"regexp": "^ghi$",
			},
			wantError: errors.New("required parameter `path` not set"),
		},
		{
			name:    "When parameters `regexp` and `search` are not set then it errors",
			files:   map[string]string{},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path": "test.txt",
			},
			wantError: errors.New("either parameter `regexp` or `search` must be set"),
		},
		{
			name:    "When parameters `regexp` and `search` are both set then it errors",
			files:   map[string]string{},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"regexp": "^ghi$",
				"search": "ghi",
			},
			wantError: errors.New("parameters `regexp` and `search` cannot bet set both"),
		},
		{
			name:    "When parameter `regexp` is invalid then it errors",
			files:   map[string]string{},
			factory: LineDeleteFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"regexp": "[a-z",
			},
			wantError: errors.New("compile parameter `regexp` to regular expression: error parsing regexp: missing closing ]: `[a-z$`"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestLineInsert_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:    "When `line` and `path` are set then it adds the line at the end of the file",
			files:   map[string]string{"test.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"line": "ttt",
				"path": "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\nttt\n"},
		},
		{
			name:    "When `path` is a glob pattern then it adds the line at the end of each matching file",
			files:   map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"line": "ttt",
				"path": "*.txt",
			},
			wantFiles: map[string]string{
				"test1.txt": "abc\ndef\n\nghi\njkl\nttt\n",
				"test2.txt": "abc\ndef\n\nghi\njkl\nttt\n",
			},
		},
		{
			name:    "When parameter `insertAt=BOF` then it adds the line at the beginning of the file",
			files:   map[string]string{"test.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"insertAt": "BOF",
				"line":     "ttt",
				"path":     "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:    "When the value of `insertAt` is invalid then it errors",
			files:   map[string]string{"test.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"insertAt": "EXO",
				"line":     "ttt",
				"path":     "test.txt",
			},
			wantError: errors.New("invalid value of parameter `insertAt` - can be one of BOF,EOF"),
		},
		{
			name:    "When the value of `insertAt` is not in upper-case then it still works",
			files:   map[string]string{"test.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"insertAt": "bOf",
				"line":     "ttt",
				"path":     "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "ttt\nabc\ndef\n\nghi\njkl\n"},
		},
		{
			name:    "When `insertAt=BOF` and the line is already present then there is no change",
			files:   map[string]string{"test.txt": testContent},
			factory: LineInsertFactory{},
			params: map[string]string{
				"insertAt": "BOF",
				"line":     "abc",
				"path":     "test.txt",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\nghi\njkl\n"},
		},
		{
			name:    "When `path` is not set then it errors",
			factory: LineInsertFactory{},
			params: map[string]string{
				"line": "jkl",
			},
			wantError: errors.New("required parameter `path` not set"),
		},
		{
			name:    "When `line` is not set then it errors",
			factory: LineInsertFactory{},
			params: map[string]string{
				"path": "test.txt",
			},
			wantError: errors.New("required parameter `line` not set"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}

func TestLineReplace_Apply(t *testing.T) {
	testCases := []testCase{
		{
			name:    "When `regexp` matches a line then it replaces the line",
			files:   map[string]string{"test.txt": testContent},
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "test.txt",
				"regexp": "^def$",
			},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:    "When `search` matches a line then it replaces the line",
			files:   map[string]string{"test.txt": testContent},
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "test.txt",
				"search": "def",
			},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:    "When `path` contains a glob pattern then it replaces the line in each matching file",
			files:   map[string]string{"test1.txt": testContent, "test2.txt": testContent},
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "*.txt",
				"search": "def",
			},
			wantFiles: map[string]string{"test1.txt": "abc\nttt\n\nghi\njkl\n", "test2.txt": "abc\nttt\n\nghi\njkl\n"},
		},
		{
			name:    "When `regexp` matches multiple lines then it replaces each line",
			files:   map[string]string{"test.txt": testContent},
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "test.txt",
				"regexp": "def|ghi",
			},
			wantFiles: map[string]string{"test.txt": "abc\nttt\n\nttt\njkl\n"},
		},
		{
			name:    "When `regexp` contains match groups then it allows its usage",
			files:   map[string]string{"test.txt": testContent},
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "${1}T${2}",
				"path":   "test.txt",
				"regexp": "(.+)h(.+)",
			},
			wantFiles: map[string]string{"test.txt": "abc\ndef\n\ngTi\njkl\n"},
		},
		{
			name:    "When `path` is not set then it errors",
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"search": "def",
			},
			wantError: errors.New("required parameter `path` not set"),
		},
		{
			name:    "When `regexp` and `search` are not set then it errors",
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line": "ttt",
				"path": "test.txt",
			},
			wantError: errors.New("either parameter `regexp` or `search` must be set"),
		},
		{
			name:    "When `regexp` and `search` are both set then it errors",
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "test.txt",
				"regexp": "^jkl$",
				"search": "def",
			},
			wantError: errors.New("parameters `regexp` and `search` cannot be set both"),
		},
		{
			name:    "When `line` is not set then it errors",
			factory: LineReplaceFactory{},
			params: map[string]string{
				"path":   "test.txt",
				"search": "def",
			},
			wantError: errors.New("required parameter `line` not set"),
		},
		{
			name:    "When the value of `regexp` is invalid then it errors",
			factory: LineReplaceFactory{},
			params: map[string]string{
				"line":   "ttt",
				"path":   "test.txt",
				"regexp": "[a-z",
			},
			wantError: errors.New("compile parameter `regexp` to regular expression: error parsing regexp: missing closing ]: `[a-z$`"),
		},
	}

	for _, tc := range testCases {
		runTestCase(t, tc)
	}
}
