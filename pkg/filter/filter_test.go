package filter_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"go.uber.org/mock/gomock"
)

type testCase struct {
	name             string
	factory          filter.Factory
	createOpts       func(*gomock.Controller) filter.CreateOptions
	params           params.Params
	repoMockFunc     func(*MockRepository)
	wantMatch        bool
	wantFactoryError string
	wantFilterError  string
}

func runTestCase(t *testing.T, tc testCase) {
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	if tc.repoMockFunc != nil {
		tc.repoMockFunc(repoMock)
	}

	var createOpts filter.CreateOptions
	if tc.createOpts != nil {
		createOpts = tc.createOpts(ctrl)
	}

	f, err := tc.factory.Create(createOpts, tc.params)
	if tc.wantFactoryError == "" {
		require.NoError(t, err)
	} else {
		require.EqualError(t, err, tc.wantFactoryError)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repoMock)
	result, err := f.Do(ctx)
	if tc.wantFilterError == "" {
		require.NoError(t, err)
		require.Equal(t, tc.wantMatch, result)
	} else {
		require.EqualError(t, err, tc.wantFilterError)
	}
}

func TestFileFactory_Create(t *testing.T) {
	opts := filter.CreateOptions{}
	fac := filter.FileFactory{}
	_, err := fac.Create(opts, map[string]any{})
	require.ErrorContains(t, err, "required parameter `paths` not set")

	_, err = fac.Create(opts, map[string]any{"op": "invalid", "paths": []string{"test.yaml"}})
	require.ErrorContains(t, err, "value of parameter `op` can be and,or not 'invalid'")
}

func TestFile_Do(t *testing.T) {
	opts := filter.CreateOptions{}
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	repoMock.EXPECT().HasFile("test.yaml").Return(true, nil).AnyTimes()
	repoMock.EXPECT().HasFile("test.json").Return(false, nil).AnyTimes()
	repoMock.EXPECT().HasFile("test.toml").Return(true, nil).AnyTimes()
	repoMock.EXPECT().HasFile("test.json5").Return(false, nil).AnyTimes()
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repoMock)

	fac := filter.FileFactory{}

	// One file, exists
	f, err := fac.Create(opts, map[string]any{"paths": []any{"test.yaml"}})
	require.NoError(t, err)
	result, err := f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// One file, missing
	f, err = fac.Create(opts, map[string]any{"paths": []any{"test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)

	// Two files, all exist, and
	f, err = fac.Create(opts, map[string]any{"op": "and", "paths": []any{"test.yaml", "test.toml"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// Two files, one missing, and
	f, err = fac.Create(opts, map[string]any{"op": "and", "paths": []any{"test.yaml", "test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)

	// Two files, one exists, or
	f, err = fac.Create(opts, map[string]any{"op": "or", "paths": []any{"test.yaml", "test.json"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.True(t, result)

	// Two files, both missing, or
	f, err = fac.Create(opts, map[string]any{"op": "or", "paths": []any{"test.json", "test.json5"}})
	require.NoError(t, err)
	result, err = f.Do(ctx)
	require.NoError(t, err)
	require.False(t, result)
}

func TestFileContentFactory_Create(t *testing.T) {
	fac := filter.FileContentFactory{}
	testCases := []testCase{
		{
			name:             "missing parameter path",
			factory:          fac,
			params:           params.Params{},
			wantFactoryError: "required parameter `path` not set",
		},
		{
			name:             "missing parameter regexp",
			factory:          fac,
			params:           params.Params{"path": "path.txt"},
			wantFactoryError: "required parameter `regexp` not set",
		},
		{
			name:             "invalid regexp",
			factory:          fac,
			params:           params.Params{"path": "path.txt", "regexp": "[a-z"},
			wantFactoryError: "compile parameter `regexp` to regular expression: error parsing regexp: missing closing ]: `[a-z`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}

func TestFileContent_Do(t *testing.T) {
	content := `abc
def
ghi
`
	fac := filter.FileContentFactory{}
	testCases := []testCase{
		{
			name:    "returns true when regexp matches content",
			factory: fac,
			params:  params.Params{"path": "test.txt", "regexp": "d?f"},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("test.txt").
					Return(content, nil).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns false when regexp does not match content",
			factory: fac,
			params:  params.Params{"path": "test.txt", "regexp": "jkl"},
			repoMockFunc: func(mr *MockRepository) {
				mr.EXPECT().
					GetFile("test.txt").
					Return(content, nil).
					Times(1)
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

func TestRepositoryFactory_Create(t *testing.T) {
	opts := filter.CreateOptions{}
	fac := filter.RepositoryFactory{}
	_, err := fac.Create(opts, map[string]any{})
	require.ErrorContains(t, err, "required parameter `host` not set")

	_, err = fac.Create(opts, map[string]any{"host": "github.com"})
	require.ErrorContains(t, err, "required parameter `owner` not set")

	_, err = fac.Create(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.Create(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.Create(opts, map[string]any{
		"host": "(github.com",
	})
	require.ErrorContains(t, err, "compile parameter `host` to regular expression: error parsing regexp: missing closing ): `^(github.com$`")

	_, err = fac.Create(opts, map[string]any{
		"host":  "github.com",
		"owner": "(wndhydrnt",
	})
	require.ErrorContains(t, err, "compile parameter `owner` to regular expression: error parsing regexp: missing closing ): `^(wndhydrnt$`")

	_, err = fac.Create(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
		"name":  "(saturn-bot",
	})
	require.ErrorContains(t, err, "compile parameter `name` to regular expression: error parsing regexp: missing closing ): `^(saturn-bot$`")
}

func TestRepository_Do(t *testing.T) {
	fac := filter.RepositoryFactory{}

	f, err := fac.Create(filter.CreateOptions{}, map[string]any{"host": "github.com", "owner": "wndhydrnt", "name": "rcmt"})
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
			hostMock := NewMockHostDetail(ctrl)
			hostMock.EXPECT().Name().Return(tc.host)
			repoMock := NewMockRepository(ctrl)
			repoMock.EXPECT().Host().Return(hostMock).AnyTimes()
			repoMock.EXPECT().Name().Return(tc.name).AnyTimes()
			repoMock.EXPECT().Owner().Return(tc.owner).AnyTimes()
			ctx := context.Background()
			ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repoMock)
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
	f := filter.NewReverse(&mockFilter{})
	result, _ := f.Do(context.Background())
	require.False(t, result)
}
