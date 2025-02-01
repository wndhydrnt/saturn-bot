package filter_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	hostmock "github.com/wndhydrnt/saturn-bot/test/mock/host"
	"go.uber.org/mock/gomock"
)

type testCase struct {
	name              string
	factory           func(filter.CreateOptions, params.Params) (filter.Filter, error)
	createOpts        func(*gomock.Controller) filter.CreateOptions
	params            params.Params
	repoMockFunc      func(*hostmock.MockRepository)
	filesInRepository map[string]string
	wantMatch         bool
	wantFactoryError  string
	wantFilterError   string
}

func runTestCase(t *testing.T, tc testCase) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repoMock := hostmock.NewMockRepository(ctrl)
	if tc.repoMockFunc != nil {
		tc.repoMockFunc(repoMock)
	}

	var createOpts filter.CreateOptions
	if tc.createOpts != nil {
		createOpts = tc.createOpts(ctrl)
	}

	f, err := tc.factory(createOpts, tc.params)
	if tc.wantFactoryError == "" {
		require.NoError(t, err)
	} else {
		require.EqualError(t, err, tc.wantFactoryError)
		return
	}

	ctx := context.Background()

	tmpDir := t.TempDir()
	for name, content := range tc.filesInRepository {
		full := filepath.Join(tmpDir, name)
		dirPrefix := filepath.Dir(full)
		err := os.MkdirAll(dirPrefix, 0755)
		require.NoError(t, err, "Creates directory of test file")
		err = os.WriteFile(full, []byte(content), 0600)
		require.NoError(t, err, "Creates the test file")
	}

	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repoMock)
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, tmpDir)
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
	_, err := fac.CreatePostClone(opts, map[string]any{})
	require.ErrorContains(t, err, "required parameter `paths` not set")

	_, err = fac.CreatePostClone(opts, map[string]any{"op": "invalid", "paths": []string{"test.yaml"}})
	require.ErrorContains(t, err, "value of parameter `op` can be and,or not 'invalid'")
}

func TestFile_Do(t *testing.T) {
	fac := filter.FileFactory{}
	testCases := []testCase{
		{
			name:    "one file, exists",
			factory: fac.CreatePostClone,
			params:  params.Params{"paths": []any{"test.yaml"}},
			filesInRepository: map[string]string{
				"test.yaml": "",
			},
			wantMatch: true,
		},
		{
			name:      "one file, missing",
			factory:   fac.CreatePostClone,
			params:    params.Params{"paths": []any{"test.json"}},
			wantMatch: false,
		},
		{
			name:    "two files, all exist, and",
			factory: fac.CreatePostClone,
			params:  params.Params{"op": "and", "paths": []any{"test.yaml", "test.toml"}},
			filesInRepository: map[string]string{
				"test.yaml": "",
				"test.toml": "",
			},
			wantMatch: true,
		},
		{
			name:    "two files, one missing, and",
			factory: fac.CreatePostClone,
			params:  params.Params{"op": "and", "paths": []any{"test.yaml", "test.json"}},
			filesInRepository: map[string]string{
				"test.yaml": "",
			},
			wantMatch: false,
		},
		{
			name:    "two files, one exists, or",
			factory: fac.CreatePostClone,
			params:  params.Params{"op": "or", "paths": []any{"test.yaml", "test.json"}},
			filesInRepository: map[string]string{
				"test.yaml": "",
			},
			wantMatch: true,
		},
		{
			name:    "two files, both missing, or",
			factory: fac.CreatePostClone,
			params:  params.Params{"op": "or", "paths": []any{"test.json", "test.json5"}},
			filesInRepository: map[string]string{
				"test.yaml": "",
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

func TestFileContentFactory_Create(t *testing.T) {
	fac := filter.FileContentFactory{}
	testCases := []testCase{
		{
			name:             "missing parameter path",
			factory:          fac.CreatePostClone,
			params:           params.Params{},
			wantFactoryError: "required parameter `path` not set",
		},
		{
			name:             "missing parameter regexp",
			factory:          fac.CreatePostClone,
			params:           params.Params{"path": "path.txt"},
			wantFactoryError: "required parameter `regexp` not set",
		},
		{
			name:             "invalid regexp",
			factory:          fac.CreatePostClone,
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
			factory: fac.CreatePostClone,
			params:  params.Params{"path": "test.txt", "regexp": "d?f"},
			filesInRepository: map[string]string{
				"test.txt": content,
			},
			wantMatch: true,
		},
		{
			name:    "returns false when regexp does not match content",
			factory: fac.CreatePostClone,
			params:  params.Params{"path": "test.txt", "regexp": "jkl"},
			filesInRepository: map[string]string{
				"test.txt": content,
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
	_, err := fac.CreatePreClone(opts, map[string]any{})
	require.ErrorContains(t, err, "required parameter `host` not set")

	_, err = fac.CreatePreClone(opts, map[string]any{"host": "github.com"})
	require.ErrorContains(t, err, "required parameter `owner` not set")

	_, err = fac.CreatePreClone(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.CreatePreClone(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
	})
	require.ErrorContains(t, err, "required parameter `name` not set")

	_, err = fac.CreatePreClone(opts, map[string]any{
		"host": "(github.com",
	})
	require.ErrorContains(t, err, "compile parameter `host` to regular expression: error parsing regexp: missing closing ): `^(github.com$`")

	_, err = fac.CreatePreClone(opts, map[string]any{
		"host":  "github.com",
		"owner": "(wndhydrnt",
	})
	require.ErrorContains(t, err, "compile parameter `owner` to regular expression: error parsing regexp: missing closing ): `^(wndhydrnt$`")

	_, err = fac.CreatePreClone(opts, map[string]any{
		"host":  "github.com",
		"owner": "wndhydrnt",
		"name":  "(saturn-bot",
	})
	require.ErrorContains(t, err, "compile parameter `name` to regular expression: error parsing regexp: missing closing ): `^(saturn-bot$`")
}

func TestRepository_Do(t *testing.T) {
	fac := filter.RepositoryFactory{}

	f, err := fac.CreatePreClone(filter.CreateOptions{}, map[string]any{"host": "github.com", "owner": "wndhydrnt", "name": "rcmt"})
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
			hostMock := hostmock.NewMockHostDetail(ctrl)
			hostMock.EXPECT().Name().Return(tc.host)
			repoMock := hostmock.NewMockRepository(ctrl)
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
