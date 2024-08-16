package filter

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	gsContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/mock"
	"go.uber.org/mock/gomock"
)

type repositoryMock struct {
	getFileResult map[string]string
	hasFileResult map[string]bool
	host          host.HostDetail
	name          string
	owner         string
}

func (r *repositoryMock) GetFile(fileName string) (string, error) {
	return r.getFileResult[fileName], nil
}

func (r *repositoryMock) HasFile(path string) (bool, error) {
	return r.hasFileResult[path], nil
}

func (r *repositoryMock) Host() host.HostDetail {
	return r.host
}

func (r *repositoryMock) Name() string {
	return r.name
}

func (r *repositoryMock) Owner() string {
	return r.owner
}

func TestFileFactory_Create(t *testing.T) {
	fac := FileFactory{}
	_, err := fac.Create(map[string]any{})
	require.ErrorContains(t, err, "required parameter `paths` not set")

	_, err = fac.Create(map[string]any{"op": "invalid", "paths": []string{"test.yaml"}})
	require.ErrorContains(t, err, "value of parameter `op` can be and,or not 'invalid'")
}

func TestFile_Do(t *testing.T) {
	repo := &repositoryMock{hasFileResult: map[string]bool{
		"test.yaml":  true,
		"test.json":  false,
		"test.toml":  true,
		"test.json5": false,
	}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	fac := FileFactory{}

	// One file, exists
	f, err := fac.Create(map[string]any{"paths": []any{"test.yaml"}})
	require.NoError(t, err)
	result, err := f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// One file, missing
	f, err = fac.Create(map[string]any{"paths": []any{"test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)

	// Two files, all exist, and
	f, err = fac.Create(map[string]any{"op": "and", "paths": []any{"test.yaml", "test.toml"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// Two files, one missing, and
	f, err = fac.Create(map[string]any{"op": "and", "paths": []any{"test.yaml", "test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)

	// Two files, one exists, or
	f, err = fac.Create(map[string]any{"op": "or", "paths": []any{"test.yaml", "test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// Two files, both missing, or
	f, err = fac.Create(map[string]any{"op": "or", "paths": []any{"test.json", "test.json5"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)
}

func TestFileContentFactory_Create(t *testing.T) {
	fac := FileContentFactory{}
	_, err := fac.Create(map[string]any{})
	require.ErrorContains(t, err, "required parameter `path` not set")

	_, err = fac.Create(map[string]any{"path": "path.txt"})
	require.ErrorContains(t, err, "required parameter `regexp` not set")

	_, err = fac.Create(map[string]any{"path": "path.txt", "regexp": "[a-z"})
	require.ErrorContains(t, err, "compile parameter `regexp` to regular expression: error parsing regexp: missing closing ]: `[a-z`")
}

func TestFileContent_Do(t *testing.T) {
	content := `abc
def
ghi
`
	repo := &repositoryMock{getFileResult: map[string]string{"test.txt": content}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	fac := FileContentFactory{}

	f, err := fac.Create(map[string]any{"path": "test.txt", "regexp": "d?f"})
	require.NoError(t, err)
	result, err := f.Do(ctx)

	require.NoError(t, err)
	require.True(t, result)

	f, err = fac.Create(map[string]any{"path": "test.txt", "regexp": "jkl"})
	require.NoError(t, err)
	result, err = f.Do(ctx)

	require.NoError(t, err)
	require.False(t, result)
}

func TestRepositoryFactory_Create(t *testing.T) {
	fac := RepositoryFactory{}
	_, err := fac.Create(map[string]any{})
	require.ErrorContains(t, err, "required parameter `host` not set")

	_, err = fac.Create(map[string]any{"host": "github.com"})
	require.ErrorContains(t, err, "required parameter `owner` not set")

	_, err = fac.Create(map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.Create(map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.Create(map[string]any{
		"host": "(github.com",
	})
	require.ErrorContains(t, err, "compile parameter `host` to regular expression: error parsing regexp: missing closing ): `^(github.com$`")

	_, err = fac.Create(map[string]any{
		"host":  "github.com",
		"owner": "(wndhydrnt",
	})
	require.ErrorContains(t, err, "compile parameter `owner` to regular expression: error parsing regexp: missing closing ): `^(wndhydrnt$`")

	_, err = fac.Create(map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
		"name":  "(saturn-bot",
	})
	require.ErrorContains(t, err, "compile parameter `name` to regular expression: error parsing regexp: missing closing ): `^(saturn-bot$`")
}

func TestRepository_Do(t *testing.T) {
	fac := RepositoryFactory{}

	f, err := fac.Create(map[string]any{"host": "github.com", "owner": "wndhydrnt", "name": "rcmt"})
	require.NoError(t, err)

	cases := []struct {
		host  string
		name  string
		owner string
		want  bool
	}{
		{
			host:  "github.com",
			owner: "wndhydrnt",
			name:  "rcmt",
			want:  true,
		},
		{
			host:  "github.com",
			owner: "prometheus",
			name:  "node_exporter",
			want:  false,
		},
		{
			host:  "github.com",
			owner: "wndhydrnt",
			name:  "rcmt-test",
			want:  false,
		},
	}

	for _, tc := range cases {
		testName := fmt.Sprintf("%s/%s/%s", tc.host, tc.owner, tc.name)
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			hostMock := mock.NewMockHostDetail(ctrl)
			hostMock.EXPECT().Name().Return(tc.host)
			repoMock := &repositoryMock{
				host:  hostMock,
				name:  tc.name,
				owner: tc.owner,
			}
			ctx := context.Background()
			ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repoMock)
			result, err := f.Do(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.want, result)
		})
	}
}

type mockFilter struct{}

func (m *mockFilter) Do(_ context.Context) (bool, error) {
	return true, nil
}

func (m *mockFilter) Name() string { return "" }

func (m *mockFilter) String() string { return "" }

func TestReverse_Do(t *testing.T) {
	f := NewReverse(&mockFilter{})
	result, _ := f.Do(context.Background())
	require.False(t, result)
}
