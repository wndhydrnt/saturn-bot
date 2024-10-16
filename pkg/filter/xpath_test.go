package filter_test

import (
	"errors"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

func TestXpathFactory_Create(t *testing.T) {
	fac := filter.XpathFactory{}
	testCases := []testCase{
		{
			name:             "errors when parameter expressions is not set",
			factory:          fac,
			params:           params.Params{"path": "pom.xml"},
			wantFactoryError: "required parameter `expressions` not set",
		},
		{
			name:             "errors when parameter expressions does not contain strings",
			factory:          fac,
			params:           params.Params{"expressions": []any{"abc", 1}, "path": "pom.xml"},
			wantFactoryError: "parameter `expressions[1]` is of type int not string",
		},
		{
			name:             "errors when parameter expressions contains an invalid XPath expression",
			factory:          fac,
			params:           params.Params{"expressions": []any{"///project"}, "path": "pom.xml"},
			wantFactoryError: "compile `expressions[0]` XPath expression: expression must evaluate to a node-set",
		},
		{
			name:             "errors when parameter path is not set",
			factory:          fac,
			params:           params.Params{"expressions": []any{"//project"}},
			wantFactoryError: "required parameter `path` not set",
		},
		{
			name:             "errors when parameter path is not of type string",
			factory:          fac,
			params:           params.Params{"expressions": []any{"//project"}, "path": 123},
			wantFactoryError: "parameter `path` is of type int not string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestXpath_Do(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8" ?>
<project>
  <dependencies>
    <dependency>
      <groupId>org.jetbrains.kotlin</groupId>
      <artifactId>kotlin-stdlib</artifactId>
      <version>2.0.0</version>
    </dependency>
    <dependency>
      <groupId>io.grpc</groupId>
      <artifactId>grpc-netty</artifactId>
      <version>1.45.4</version>
    </dependency>
  </dependencies>
</project>
`
	fac := filter.XpathFactory{}
	testCases := []testCase{
		{
			name:    "returns true when expression matches a node",
			factory: fac,
			params: params.Params{
				"expressions": []any{`/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]//version[starts-with(text(), "2")]`},
				"path":        "pom.xml",
			},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("pom.xml").
					Return(content, nil)
			},
			wantMatch: true,
		},
		{
			name:    "returns false when expression does not match any nodes",
			factory: fac,
			params: params.Params{
				"expressions": []any{`/project//dependencies//dependency[artifactId/text()="kotlin-bom"]`},
				"path":        "pom.xml",
			},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("pom.xml").
					Return(content, nil)
			},
			wantMatch: false,
		},
		{
			name:    "returns false when the repository does not contain the file",
			factory: fac,
			params: params.Params{
				"expressions": []any{`/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`},
				"path":        "other.xml",
			},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("other.xml").
					Return("", host.ErrFileNotFound)
			},
			wantMatch: false,
		},
		{
			name:    "errors when the download of the file fails",
			factory: fac,
			params: params.Params{
				"expressions": []any{`/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`},
				"path":        "failure.xml",
			},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("failure.xml").
					Return("", errors.New("internal server error"))
			},
			wantFilterError: "download file from repository: internal server error",
		},
		{
			name:    "returns false when the file contains invalid XML",
			factory: fac,
			params: params.Params{
				"expressions": []any{`/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`},
				"path":        "invalid.xml",
			},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("invalid.xml").
					Return("<<project>invalid</project>", nil)
			},
			wantMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
