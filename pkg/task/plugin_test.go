package task_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	proto "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	gsContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/mock/gomock"
)

func setupRepoPluginTest(ctrl *gomock.Controller) (repoMock *MockRepository, payload *proto.Repository) {
	repoMock = NewMockRepository(ctrl)
	repoMock.EXPECT().FullName().Return("git.localhost/unit/test").AnyTimes()
	repoMock.EXPECT().CloneUrlHttp().Return("https://git.localhost/unit/test.git").AnyTimes()
	repoMock.EXPECT().CloneUrlSsh().Return("git@git.localhost/unit/test.git").AnyTimes()
	repoMock.EXPECT().WebUrl().Return("https://git.localhost/unit/test").AnyTimes()

	payload = &proto.Repository{
		FullName:     "git.localhost/unit/test",
		CloneUrlHttp: "https://git.localhost/unit/test.git",
		CloneUrlSsh:  "git@git.localhost/unit/test.git",
		WebUrl:       "https://git.localhost/unit/test",
	}
	return repoMock, payload
}

func TestPluginAction_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	runData := map[string]string{"a": "1"}
	provider := NewMockProvider(ctrl)
	provider.EXPECT().ExecuteActions(&proto.ExecuteActionsRequest{
		Context: &proto.Context{
			RunData:    runData,
			Repository: payload,
		},
		Path: "/tmp",
	}).Return(&proto.ExecuteActionsResponse{
		RunData: map[string]string{
			"a": "2", // Should update key "a" because it already exists.
			"b": "2", // Should be added because key "b" does not exist.
		},
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, gsContext.RunDataKey{}, runData)
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	pa := &task.PluginAction{Provider: provider}
	err := pa.Apply(ctx)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "2", "b": "2"}, runData)
}

func TestPluginAction_Apply_WithPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	provider := NewMockProvider(ctrl)
	provider.EXPECT().ExecuteActions(&proto.ExecuteActionsRequest{
		Context: &proto.Context{
			RunData:     make(map[string]string),
			PullRequest: &proto.PullRequest{Number: 123, WebUrl: "https://git.localhost/unit/test/pulls/123"},
			Repository:  payload,
		},
		Path: "/tmp",
	}).Return(&proto.ExecuteActionsResponse{}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, gsContext.PullRequestKey{}, host.PullRequest{
		Number: 123,
		WebURL: "https://git.localhost/unit/test/pulls/123",
	})
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	pa := &task.PluginAction{Provider: provider}
	err := pa.Apply(ctx)

	require.NoError(t, err)
}

func TestPluginFilter_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	runData := map[string]string{"a": "1"}
	provider := NewMockProvider(ctrl)
	provider.EXPECT().ExecuteFilters(&proto.ExecuteFiltersRequest{
		Context: &proto.Context{
			Repository: payload,
			RunData:    runData,
		},
	}).Return(&proto.ExecuteFiltersResponse{
		Match: true,
		RunData: map[string]string{
			"a": "2", // Should update key "a" because it already exists.
			"b": "2", // Should be added to run data because key "b" does not exist.
		},
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)
	ctx = context.WithValue(ctx, gsContext.RunDataKey{}, runData)

	pf := &task.PluginFilter{Provider: provider}
	match, err := pf.Do(ctx)

	require.NoError(t, err)
	require.True(t, match)
	require.Equal(t, map[string]string{"a": "2", "b": "2"}, runData)
}

func TestPluginFilter_Do_WithPullRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo, payload := setupRepoPluginTest(ctrl)
	provider := NewMockProvider(ctrl)
	provider.EXPECT().ExecuteFilters(&proto.ExecuteFiltersRequest{
		Context: &proto.Context{
			RunData:     make(map[string]string),
			PullRequest: &proto.PullRequest{Number: 123, WebUrl: "https://git.localhost/unit/test/pulls/123"},
			Repository:  payload,
		},
	}).Return(&proto.ExecuteFiltersResponse{
		Match: true,
	}, nil)
	ctx := context.Background()
	ctx = context.WithValue(ctx, gsContext.CheckoutPath{}, "/tmp")
	ctx = context.WithValue(ctx, gsContext.PullRequestKey{}, host.PullRequest{
		Number: 123,
		WebURL: "https://git.localhost/unit/test/pulls/123",
	})
	ctx = context.WithValue(ctx, gsContext.RepositoryKey{}, repo)

	pf := &task.PluginFilter{Provider: provider}
	match, err := pf.Do(ctx)

	require.NoError(t, err)
	require.True(t, match)
}
