package plugin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/plugin"
	hostmock "github.com/wndhydrnt/saturn-bot/test/mock/host"
	pluginmock "github.com/wndhydrnt/saturn-bot/test/mock/plugin"
	"go.uber.org/mock/gomock"
)

func setupRepoPluginTest(ctrl *gomock.Controller) (repoMock *hostmock.MockRepository, payload *protoV1.Repository) {
	repoMock = hostmock.NewMockRepository(ctrl)
	repoMock.EXPECT().FullName().Return("git.localhost/unit/test").AnyTimes()
	repoMock.EXPECT().CloneUrlHttp().Return("https://git.localhost/unit/test.git").AnyTimes()
	repoMock.EXPECT().CloneUrlSsh().Return("git@git.localhost/unit/test.git").AnyTimes()
	repoMock.EXPECT().WebUrl().Return("https://git.localhost/unit/test").AnyTimes()

	payload = &protoV1.Repository{
		FullName:     "git.localhost/unit/test",
		CloneUrlHttp: "https://git.localhost/unit/test.git",
		CloneUrlSsh:  "git@git.localhost/unit/test.git",
		WebUrl:       "https://git.localhost/unit/test",
	}
	return repoMock, payload
}

func TestPlugin_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	runData := map[string]string{"a": "1"}
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteActions(&protoV1.ExecuteActionsRequest{
		Context: &protoV1.Context{
			RunData:    runData,
			Repository: payload,
		},
		Path: "/tmp",
	}).Return(&protoV1.ExecuteActionsResponse{
		RunData: map[string]string{
			"a": "2", // Should update key "a" because it already exists.
			"b": "2", // Should be added because key "b" does not exist.
		},
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = sbcontext.WithRunData(ctx, runData)
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	pa := &plugin.Plugin{Provider: provider}
	err := pa.Apply(ctx)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "2", "b": "2"}, sbcontext.RunData(ctx))
}

func TestPlugin_Apply_WithPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteActions(&protoV1.ExecuteActionsRequest{
		Context: &protoV1.Context{
			RunData:     make(map[string]string),
			PullRequest: &protoV1.PullRequest{Number: 123, WebUrl: "https://git.localhost/unit/test/pulls/123"},
			Repository:  payload,
		},
		Path: "/tmp",
	}).Return(&protoV1.ExecuteActionsResponse{}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, sbcontext.PullRequestKey{}, host.PullRequest{
		Number: 123,
		WebURL: "https://git.localhost/unit/test/pulls/123",
	})
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	pa := &plugin.Plugin{Provider: provider}
	err := pa.Apply(ctx)

	require.NoError(t, err)
}

func TestPlugin_Apply_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	errMsg := "exception in plugin"
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteActions(&protoV1.ExecuteActionsRequest{
		Context: &protoV1.Context{
			RunData:    map[string]string{},
			Repository: payload,
		},
		Path: "/tmp",
	}).Return(&protoV1.ExecuteActionsResponse{
		Error: &errMsg,
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	pa := &plugin.Plugin{Provider: provider}
	err := pa.Apply(ctx)

	require.ErrorContains(t, err, "execute action: exception in plugin")
}

func TestPlugin_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	runData := map[string]string{"a": "1"}
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteFilters(&protoV1.ExecuteFiltersRequest{
		Context: &protoV1.Context{
			Repository: payload,
			RunData:    runData,
		},
	}).Return(&protoV1.ExecuteFiltersResponse{
		Match: true,
		RunData: map[string]string{
			"a": "2", // Should update key "a" because it already exists.
			"b": "2", // Should be added to run data because key "b" does not exist.
		},
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)
	ctx = sbcontext.WithRunData(ctx, runData)

	pf := &plugin.Plugin{Provider: provider}
	match, err := pf.Do(ctx)

	require.NoError(t, err)
	require.True(t, match)
	require.Equal(t, map[string]string{"a": "2", "b": "2"}, sbcontext.RunData(ctx))
}

func TestPlugin_Do_WithPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteFilters(&protoV1.ExecuteFiltersRequest{
		Context: &protoV1.Context{
			RunData:     make(map[string]string),
			PullRequest: &protoV1.PullRequest{Number: 123, WebUrl: "https://git.localhost/unit/test/pulls/123"},
			Repository:  payload,
		},
	}).Return(&protoV1.ExecuteFiltersResponse{
		Match: true,
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, sbcontext.PullRequestKey{}, host.PullRequest{
		Number: 123,
		WebURL: "https://git.localhost/unit/test/pulls/123",
	})
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	pf := &plugin.Plugin{Provider: provider}
	match, err := pf.Do(ctx)

	require.NoError(t, err)
	require.True(t, match)
}

func TestPlugin_Do_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	errMsg := "exception in plugin"
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().ExecuteFilters(&protoV1.ExecuteFiltersRequest{
		Context: &protoV1.Context{
			Repository: payload,
			RunData:    map[string]string{},
		},
	}).Return(&protoV1.ExecuteFiltersResponse{
		Error: &errMsg,
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	pf := &plugin.Plugin{Provider: provider}
	match, err := pf.Do(ctx)

	require.ErrorContains(t, err, "execute filter: exception in plugin")
	require.False(t, match)
}

func TestPlugin_Start_Stop(t *testing.T) {
	opts := plugin.StartOptions{Config: map[string]string{"message": "Hello World"}}
	ctrl := gomock.NewController(t)
	provider := pluginmock.NewMockProvider(ctrl)
	provider.EXPECT().
		GetPlugin(&protoV1.GetPluginRequest{Config: opts.Config}).
		Return(&protoV1.GetPluginResponse{Name: "unittest", Priority: 10}, nil)
	provider.EXPECT().
		Shutdown(&protoV1.ShutdownRequest{}).
		Return(&protoV1.ShutdownResponse{}, nil)

	p := &plugin.Plugin{Provider: provider}
	err := p.Start(opts)
	p.Stop()

	require.NoError(t, err)
	require.Equal(t, "unittest", p.Name)
}

func TestPlugin_String(t *testing.T) {
	a := &plugin.Plugin{Name: "unittest"}
	require.Equal(t, "plugin(name=unittest)", a.String())
}

func TestGetPluginExec(t *testing.T) {
	type args struct {
		path       string
		javaExec   string
		pythonExec string
	}
	type testCase struct {
		args   args
		result plugin.ExecOptions
	}

	testCases := []testCase{
		{
			args: args{
				path: "/opt/plugin",
			},
			result: plugin.ExecOptions{
				Executable: "/opt/plugin",
			},
		},
		{
			args: args{
				path: "/opt/plugin.jar",
			},
			result: plugin.ExecOptions{
				Args:       []string{"-jar", "/opt/plugin.jar"},
				Executable: "java",
			},
		},
		{
			args: args{
				path:     "/opt/plugin.jar",
				javaExec: "/opt/java21/bin/java",
			},
			result: plugin.ExecOptions{
				Args:       []string{"-jar", "/opt/plugin.jar"},
				Executable: "/opt/java21/bin/java",
			},
		},
		{
			args: args{
				path: "/opt/plugin.py",
			},
			result: plugin.ExecOptions{
				Args:       []string{"/opt/plugin.py"},
				Executable: "python",
			},
		},
		{
			args: args{
				path:       "/opt/plugin.py",
				pythonExec: "/var/venv/bin/python",
			},
			result: plugin.ExecOptions{
				Args:       []string{"/opt/plugin.py"},
				Executable: "/var/venv/bin/python",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.args), func(t *testing.T) {
			result := plugin.GetPluginExec(tc.args.path, tc.args.javaExec, tc.args.pythonExec)
			require.Equal(t, tc.result, result)
		})
	}
}
