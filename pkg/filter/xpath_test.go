package filter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"go.uber.org/mock/gomock"
)

func TestXpathFactory_Create(t *testing.T) {
	fac := filter.XpathFactory{}

	_, err := fac.Create(map[string]any{
		"path": "pom.xml",
	})
	require.EqualError(t, err, "required parameter `expression` not set")

	_, err = fac.Create(map[string]any{
		"expression": 123,
		"path":       "pom.xml",
	})
	require.EqualError(t, err, "parameter `expression` is of type int not string")

	_, err = fac.Create(map[string]any{
		"expression": "///project",
		"path":       "pom.xml",
	})
	require.EqualError(t, err, "invalid XPath expression: expression must evaluate to a node-set")

	_, err = fac.Create(map[string]any{
		"expression": "//project",
	})
	require.EqualError(t, err, "required parameter `path` not set")

	_, err = fac.Create(map[string]any{
		"expression": "//project",
		"path":       123,
	})
	require.EqualError(t, err, "parameter `path` is of type int not string")
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
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	repoMock.EXPECT().
		GetFile("pom.xml").
		Return(content, nil).
		Times(2)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repoMock)

	fac := filter.XpathFactory{}

	f, err := fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]//version[starts-with(text(), "2")]`,
		"path":       "pom.xml",
	})
	require.NoError(t, err)
	result, err := f.Do(ctx)

	require.NoError(t, err)
	require.True(t, result)

	f, err = fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-bom"]`,
		"path":       "pom.xml",
	})
	require.NoError(t, err)
	result, err = f.Do(ctx)

	require.NoError(t, err)
	require.False(t, result)

	f, err = fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`,
		"path":       "pom.xml",
	})
	require.NoError(t, err)
	_, err = f.Do(context.Background())
	require.EqualError(t, err, "context does not contain a repository")

	repoMock.EXPECT().
		GetFile("other.xml").
		Return("", host.ErrFileNotFound).
		Times(1)
	f, err = fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`,
		"path":       "other.xml",
	})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)

	repoMock.EXPECT().
		GetFile("failure.xml").
		Return("", errors.New("internal server error")).
		Times(1)
	f, err = fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`,
		"path":       "failure.xml",
	})
	require.NoError(t, err)
	_, err = f.Do(ctx)
	require.EqualError(t, err, "download file from repository: internal server error")

	repoMock.EXPECT().
		GetFile("invalid.xml").
		Return("<<project>invalid</project>", nil).
		Times(1)
	f, err = fac.Create(map[string]any{
		"expression": `/project//dependencies//dependency[artifactId/text()="kotlin-stdlib"]`,
		"path":       "invalid.xml",
	})
	require.NoError(t, err)
	_, err = f.Do(ctx)

	require.EqualError(t, err, "parse XML document: XML syntax error on line 1: expected element name after <")
}
