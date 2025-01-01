package filter_test

import (
	"fmt"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"go.uber.org/mock/gomock"
)

type MockGitLabSearcher struct {
	host.Host

	groupID any
	query   string

	projectIDs []int64
}

func (m *MockGitLabSearcher) SearchCode(groupID any, query string) ([]int64, error) {
	if m.groupID == groupID && m.query == query {
		return m.projectIDs, nil
	}

	return nil, fmt.Errorf("mock received unexpected call: gitlabGroupID=%v, query=%v", groupID, query)
}

func TestGitlabCodeSearch_Do(t *testing.T) {
	testCases := []testCase{
		{
			name:    "returns true when search result contains project ID",
			factory: filter.GitlabCodeSearchFactory{}.CreatePreClone,
			createOpts: func(ctrl *gomock.Controller) filter.CreateOptions {
				m := &MockGitLabSearcher{projectIDs: []int64{10}, query: "func"}
				return filter.CreateOptions{Hosts: []host.Host{m}}
			},
			params: params.Params{"query": "func"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					ID().
					Return(int64(10)).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "returns false when search result does not contain project ID",
			factory: filter.GitlabCodeSearchFactory{}.CreatePreClone,
			createOpts: func(ctrl *gomock.Controller) filter.CreateOptions {
				m := &MockGitLabSearcher{projectIDs: []int64{20}, query: "func"}
				return filter.CreateOptions{Hosts: []host.Host{m}}
			},
			params: params.Params{"query": "func"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					ID().
					Return(int64(10)).
					Times(1)
			},
			wantMatch: false,
		},
		{
			name:    "limits query by group ID of type int if set",
			factory: filter.GitlabCodeSearchFactory{}.CreatePreClone,
			createOpts: func(ctrl *gomock.Controller) filter.CreateOptions {
				m := &MockGitLabSearcher{
					groupID:    23,
					query:      "func",
					projectIDs: []int64{10},
				}
				return filter.CreateOptions{Hosts: []host.Host{m}}
			},
			params: params.Params{"groupID": 23, "query": "func"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					ID().
					Return(int64(10)).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:    "limits query by group ID of type string if set",
			factory: filter.GitlabCodeSearchFactory{}.CreatePreClone,
			createOpts: func(ctrl *gomock.Controller) filter.CreateOptions {
				m := &MockGitLabSearcher{
					groupID:    "unit/test",
					query:      "func",
					projectIDs: []int64{10},
				}
				return filter.CreateOptions{Hosts: []host.Host{m}}
			},
			params: params.Params{"groupID": "unit/test", "query": "func"},
			repoMockFunc: func(repoMock *MockRepository) {
				repoMock.EXPECT().
					ID().
					Return(int64(10)).
					Times(1)
			},
			wantMatch: true,
		},
		{
			name:             "errors if parameter query is not set",
			factory:          filter.GitlabCodeSearchFactory{}.CreatePreClone,
			params:           params.Params{},
			wantFactoryError: "required parameter `query` not set",
		},
		{
			name:    "errors if parameter groupID is neither int nor string",
			factory: filter.GitlabCodeSearchFactory{}.CreatePreClone,
			createOpts: func(ctrl *gomock.Controller) filter.CreateOptions {
				m := &MockGitLabSearcher{}
				return filter.CreateOptions{Hosts: []host.Host{m}}
			},
			params:           params.Params{"groupID": 1.0, "query": "func"},
			wantFactoryError: "parameter `groupID` must be either of type int or string",
		},
		{
			name:             "errors if GitLab is not configured",
			factory:          filter.GitlabCodeSearchFactory{}.CreatePreClone,
			params:           params.Params{"query": "func"},
			wantFactoryError: "required host for GitLab not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
