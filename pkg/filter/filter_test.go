package filter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	gsContext "github.com/wndhydrnt/saturn-sync/pkg/context"
)

type repositoryMock struct {
	getFileResult     map[string]string
	getFullNameResult string
	hasFileResult     map[string]bool
}

func (r *repositoryMock) GetFile(fileName string) (string, error) {
	return r.getFileResult[fileName], nil
}

func (r *repositoryMock) FullName() string {
	return r.getFullNameResult
}

func (r *repositoryMock) HasFile(path string) (bool, error) {
	return r.hasFileResult[path], nil
}

func TestFileExists_Do(t *testing.T) {
	repo := &repositoryMock{hasFileResult: map[string]bool{"test.yaml": true, "test.json": false}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	f := NewFile("test.yaml")
	result, err := f.Do(ctx)

	require.NoError(t, err)
	require.True(t, result)

	f = NewFile("test.json")
	result, err = f.Do(ctx)

	require.NoError(t, err)
	require.False(t, result)
}

func TestFileContainsLine_Do(t *testing.T) {
	content := `abc
def
ghi
`
	repo := &repositoryMock{getFileResult: map[string]string{"test.txt": content}}
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	f, err := NewFileContent("test.txt", "d?f")
	require.NoError(t, err)
	result, err := f.Do(ctx)

	require.NoError(t, err)
	require.True(t, result)

	f, err = NewFileContent("test.txt", "jkl")
	require.NoError(t, err)
	result, err = f.Do(ctx)

	require.NoError(t, err)
	require.False(t, result)
}

func TestRepositoryName_Do(t *testing.T) {
	f, err := NewRepositoryName([]string{"https://github.com/wndhydrnt/rcmt.git"})
	require.NoError(t, err)

	cases := []struct {
		toMatch string
		want    bool
	}{
		{
			toMatch: "github.com/wndhydrnt/rcmt",
			want:    true,
		},
		{
			toMatch: "github.com/prometheus/node_exporter",
			want:    false,
		},
		{
			toMatch: "github.com/wndhydrnt/rcmt-test",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.toMatch, func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, &repositoryMock{getFullNameResult: tc.toMatch})
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
