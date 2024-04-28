// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/git/git.go
//
// Generated by this command:
//
//	mockgen -package mock -source pkg/git/git.go
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	host "github.com/wndhydrnt/saturn-bot/pkg/host"
	gomock "go.uber.org/mock/gomock"
)

// MockGitClient is a mock of GitClient interface.
type MockGitClient struct {
	ctrl     *gomock.Controller
	recorder *MockGitClientMockRecorder
}

// MockGitClientMockRecorder is the mock recorder for MockGitClient.
type MockGitClientMockRecorder struct {
	mock *MockGitClient
}

// NewMockGitClient creates a new mock instance.
func NewMockGitClient(ctrl *gomock.Controller) *MockGitClient {
	mock := &MockGitClient{ctrl: ctrl}
	mock.recorder = &MockGitClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitClient) EXPECT() *MockGitClientMockRecorder {
	return m.recorder
}

// CommitChanges mocks base method.
func (m *MockGitClient) CommitChanges(msg string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitChanges", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// CommitChanges indicates an expected call of CommitChanges.
func (mr *MockGitClientMockRecorder) CommitChanges(msg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitChanges", reflect.TypeOf((*MockGitClient)(nil).CommitChanges), msg)
}

// Execute mocks base method.
func (m *MockGitClient) Execute(arg ...string) (string, string, error) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Execute", varargs...)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Execute indicates an expected call of Execute.
func (mr *MockGitClientMockRecorder) Execute(arg ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockGitClient)(nil).Execute), arg...)
}

// HasLocalChanges mocks base method.
func (m *MockGitClient) HasLocalChanges() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasLocalChanges")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasLocalChanges indicates an expected call of HasLocalChanges.
func (mr *MockGitClientMockRecorder) HasLocalChanges() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasLocalChanges", reflect.TypeOf((*MockGitClient)(nil).HasLocalChanges))
}

// HasRemoteChanges mocks base method.
func (m *MockGitClient) HasRemoteChanges(branchName string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasRemoteChanges", branchName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasRemoteChanges indicates an expected call of HasRemoteChanges.
func (mr *MockGitClientMockRecorder) HasRemoteChanges(branchName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasRemoteChanges", reflect.TypeOf((*MockGitClient)(nil).HasRemoteChanges), branchName)
}

// Prepare mocks base method.
func (m *MockGitClient) Prepare(repo host.Repository, retry bool) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prepare", repo, retry)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Prepare indicates an expected call of Prepare.
func (mr *MockGitClientMockRecorder) Prepare(repo, retry any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*MockGitClient)(nil).Prepare), repo, retry)
}

// Push mocks base method.
func (m *MockGitClient) Push(branchName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Push", branchName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Push indicates an expected call of Push.
func (mr *MockGitClientMockRecorder) Push(branchName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockGitClient)(nil).Push), branchName)
}

// UpdateTaskBranch mocks base method.
func (m *MockGitClient) UpdateTaskBranch(branchName string, forceRebase bool, repo host.Repository) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateTaskBranch", branchName, forceRebase, repo)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateTaskBranch indicates an expected call of UpdateTaskBranch.
func (mr *MockGitClientMockRecorder) UpdateTaskBranch(branchName, forceRebase, repo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateTaskBranch", reflect.TypeOf((*MockGitClient)(nil).UpdateTaskBranch), branchName, forceRebase, repo)
}
