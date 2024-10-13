package filter_test

import (
	"errors"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

func TestJsonPath_Do(t *testing.T) {
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
			name:    "returns true when JSONPath expression matches",
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$.dependencies["@commitlint/cli"]`, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns true when JSONPath expression contains match filter",
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$[?(@['@commitlint/cli'] == '19.5.0')]`, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return(packageJson, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns false when JSONPath expression doesn't match",
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$.dependencies["other"]`, "path": "package.json"},
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
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$.dependencies`, "path": "other.json"},
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
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$.dependencies`, "path": "package.json"},
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
			factory: filter.JsonPathFactory{},
			params:  params.Params{"expression": `$.dependencies`, "path": "package.json"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					GetFile("package.json").
					Return("{{}", nil).
					Times(1)
			},
			wantFilterError: "decode file for JSONPath filter: yaml: line 1: did not find expected ',' or '}'",
		},
		{
			name:             "factory errors when param expression is not set",
			factory:          filter.JsonPathFactory{},
			params:           params.Params{"path": "package.json"},
			wantFactoryError: "required parameter `expression` not set",
		},
		{
			name:             "factory errors when param expression has wrong type",
			factory:          filter.JsonPathFactory{},
			params:           params.Params{"expression": 123, "path": "package.json"},
			wantFactoryError: "parameter `expression` is of type int not string",
		},
		{
			name:             "factory errors when value of param expression is an invalid JSONPath expression",
			factory:          filter.JsonPathFactory{},
			params:           params.Params{"expression": "$$.dependencies", "path": "package.json"},
			wantFactoryError: "parse JSONPath expression: parse error at 3 in $$.dependencies",
		},
		{
			name:             "factory errors when param path is not set",
			factory:          filter.JsonPathFactory{},
			params:           params.Params{"expression": "$.dependencies"},
			wantFactoryError: "required parameter `path` not set",
		},
		{
			name:             "factory errors when param path has wrong type",
			factory:          filter.JsonPathFactory{},
			params:           params.Params{"expression": "$.dependencies", "path": 123},
			wantFactoryError: "parameter `path` is of type int not string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
