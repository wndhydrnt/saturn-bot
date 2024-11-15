package filter_test

import (
	"errors"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

func TestJq_Do(t *testing.T) {
	packageJson := `{
  "name": "docker-commitlint",
  "version": "1.0.0",
  "description": "A Docker container that contains commitlint and plugins",
  "dependencies": {
    "@commitlint/cli": "19.5.0",
    "@commitlint/config-conventional": "19.5.0"
  }
}`

	testCases := []testCase{
		{
			name:    "returns true when jq expression matches",
			factory: filter.JqFactory{},
			params:  params.Params{"expressions": []any{`.dependencies["@commitlint/cli"]`}, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns true when multiple jq expressions match",
			factory: filter.JqFactory{},
			params: params.Params{
				"expressions": []any{`.dependencies["@commitlint/cli"]`, `.name == "docker-commitlint"`},
				"path":        "package.json",
			},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns true when jq expression contains match filter",
			factory: filter.JqFactory{},
			params:  params.Params{"expressions": []any{`.dependencies["@commitlint/cli"] == "19.5.0"`}, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns false when jq expression doesn't match",
			factory: filter.JqFactory{},
			params: params.Params{
				"expressions": []any{`.dependencies["@commitlint/cli"]`, `.version == "2.0.0"`},
				"path":        "package.json",
			},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: false,
		},
		{
			name:    "returns false when field in jq expression doesn't exist in content",
			factory: filter.JqFactory{},
			params: params.Params{
				"expressions": []any{`.summary`},
				"path":        "package.json",
			},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: false,
		},
		{
			name:    "returns false when file doesn't exist in repository",
			factory: filter.JqFactory{},
			params:  params.Params{"expressions": []any{`.dependencies`}, "path": "other.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("other.json").
					Return("", host.ErrFileNotFound).
					Times(1)
			},
			wantMatch: false,
		},
		{
			name:    "errors when download of file fails",
			factory: filter.JqFactory{},
			params:  params.Params{"expressions": []any{`.dependencies`}, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return("", errors.New("internal server error")).
					Times(1)
			},
			wantFilterError: "download file from repository: internal server error",
		},
		{
			name:    "errors when file contains invalid content",
			factory: filter.JqFactory{},
			params:  params.Params{"expressions": []any{`.dependencies`}, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return("{{}", nil).
					Times(1)
			},
			wantFilterError: "decode file for jq filter: yaml: line 1: did not find expected ',' or '}'",
		},
		{
			name:             "factory errors when param expressions is not set",
			factory:          filter.JqFactory{},
			params:           params.Params{"path": "package.json"},
			wantFactoryError: "required parameter `expressions` not set",
		},
		{
			name:             "factory errors when param expressions has wrong type",
			factory:          filter.JqFactory{},
			params:           params.Params{"expressions": 123, "path": "package.json"},
			wantFactoryError: "parameter `expressions` is of type int not slice",
		},
		{
			name:             "factory errors when value of param expressions is an invalid jq expression",
			factory:          filter.JqFactory{},
			params:           params.Params{"expressions": []any{"$$.dependencies"}, "path": "package.json"},
			wantFactoryError: "parse `expressions[0]` jq expression: unexpected token \"$\"",
		},
		{
			name:             "factory errors when param path is not set",
			factory:          filter.JqFactory{},
			params:           params.Params{"expressions": []any{".dependencies"}},
			wantFactoryError: "required parameter `path` not set",
		},
		{
			name:             "factory errors when param path has wrong type",
			factory:          filter.JqFactory{},
			params:           params.Params{"expressions": []any{".dependencies"}, "path": 123},
			wantFactoryError: "parameter `path` is of type int not string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
