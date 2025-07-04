// Code generated by MockGen. DO NOT EDIT.
// Source: ../../../pkg/host/host.go
//
// Generated by this command:
//
//	mockgen -package host -source ../../../pkg/host/host.go -destination host.gen.go
//

// Package host is a generated GoMock package.
package host

import (
	json "encoding/json"
	iter "iter"
	reflect "reflect"
	time "time"

	host "github.com/wndhydrnt/saturn-bot/pkg/host"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
	isgomock struct{}
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// BaseBranch mocks base method.
func (m *MockRepository) BaseBranch() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BaseBranch")
	ret0, _ := ret[0].(string)
	return ret0
}

// BaseBranch indicates an expected call of BaseBranch.
func (mr *MockRepositoryMockRecorder) BaseBranch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BaseBranch", reflect.TypeOf((*MockRepository)(nil).BaseBranch))
}

// CanMergePullRequest mocks base method.
func (m *MockRepository) CanMergePullRequest(pr *host.PullRequest) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanMergePullRequest", pr)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CanMergePullRequest indicates an expected call of CanMergePullRequest.
func (mr *MockRepositoryMockRecorder) CanMergePullRequest(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanMergePullRequest", reflect.TypeOf((*MockRepository)(nil).CanMergePullRequest), pr)
}

// CloneUrlHttp mocks base method.
func (m *MockRepository) CloneUrlHttp() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloneUrlHttp")
	ret0, _ := ret[0].(string)
	return ret0
}

// CloneUrlHttp indicates an expected call of CloneUrlHttp.
func (mr *MockRepositoryMockRecorder) CloneUrlHttp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloneUrlHttp", reflect.TypeOf((*MockRepository)(nil).CloneUrlHttp))
}

// CloneUrlSsh mocks base method.
func (m *MockRepository) CloneUrlSsh() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloneUrlSsh")
	ret0, _ := ret[0].(string)
	return ret0
}

// CloneUrlSsh indicates an expected call of CloneUrlSsh.
func (mr *MockRepositoryMockRecorder) CloneUrlSsh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloneUrlSsh", reflect.TypeOf((*MockRepository)(nil).CloneUrlSsh))
}

// ClosePullRequest mocks base method.
func (m *MockRepository) ClosePullRequest(msg string, pr *host.PullRequest) (*host.PullRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClosePullRequest", msg, pr)
	ret0, _ := ret[0].(*host.PullRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ClosePullRequest indicates an expected call of ClosePullRequest.
func (mr *MockRepositoryMockRecorder) ClosePullRequest(msg, pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClosePullRequest", reflect.TypeOf((*MockRepository)(nil).ClosePullRequest), msg, pr)
}

// CreatePullRequest mocks base method.
func (m *MockRepository) CreatePullRequest(branch string, data host.PullRequestData) (*host.PullRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePullRequest", branch, data)
	ret0, _ := ret[0].(*host.PullRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePullRequest indicates an expected call of CreatePullRequest.
func (mr *MockRepositoryMockRecorder) CreatePullRequest(branch, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePullRequest", reflect.TypeOf((*MockRepository)(nil).CreatePullRequest), branch, data)
}

// CreatePullRequestComment mocks base method.
func (m *MockRepository) CreatePullRequestComment(body string, pr *host.PullRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePullRequestComment", body, pr)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePullRequestComment indicates an expected call of CreatePullRequestComment.
func (mr *MockRepositoryMockRecorder) CreatePullRequestComment(body, pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePullRequestComment", reflect.TypeOf((*MockRepository)(nil).CreatePullRequestComment), body, pr)
}

// DeleteBranch mocks base method.
func (m *MockRepository) DeleteBranch(pr *host.PullRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBranch", pr)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBranch indicates an expected call of DeleteBranch.
func (mr *MockRepositoryMockRecorder) DeleteBranch(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBranch", reflect.TypeOf((*MockRepository)(nil).DeleteBranch), pr)
}

// DeletePullRequestComment mocks base method.
func (m *MockRepository) DeletePullRequestComment(comment host.PullRequestComment, pr *host.PullRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePullRequestComment", comment, pr)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePullRequestComment indicates an expected call of DeletePullRequestComment.
func (mr *MockRepositoryMockRecorder) DeletePullRequestComment(comment, pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePullRequestComment", reflect.TypeOf((*MockRepository)(nil).DeletePullRequestComment), comment, pr)
}

// FindPullRequest mocks base method.
func (m *MockRepository) FindPullRequest(branch string) (*host.PullRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindPullRequest", branch)
	ret0, _ := ret[0].(*host.PullRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindPullRequest indicates an expected call of FindPullRequest.
func (mr *MockRepositoryMockRecorder) FindPullRequest(branch any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindPullRequest", reflect.TypeOf((*MockRepository)(nil).FindPullRequest), branch)
}

// FullName mocks base method.
func (m *MockRepository) FullName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FullName")
	ret0, _ := ret[0].(string)
	return ret0
}

// FullName indicates an expected call of FullName.
func (mr *MockRepositoryMockRecorder) FullName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FullName", reflect.TypeOf((*MockRepository)(nil).FullName))
}

// GetPullRequestBody mocks base method.
func (m *MockRepository) GetPullRequestBody(pr *host.PullRequest) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPullRequestBody", pr)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPullRequestBody indicates an expected call of GetPullRequestBody.
func (mr *MockRepositoryMockRecorder) GetPullRequestBody(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPullRequestBody", reflect.TypeOf((*MockRepository)(nil).GetPullRequestBody), pr)
}

// HasSuccessfulPullRequestBuild mocks base method.
func (m *MockRepository) HasSuccessfulPullRequestBuild(pr *host.PullRequest) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasSuccessfulPullRequestBuild", pr)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasSuccessfulPullRequestBuild indicates an expected call of HasSuccessfulPullRequestBuild.
func (mr *MockRepositoryMockRecorder) HasSuccessfulPullRequestBuild(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasSuccessfulPullRequestBuild", reflect.TypeOf((*MockRepository)(nil).HasSuccessfulPullRequestBuild), pr)
}

// Host mocks base method.
func (m *MockRepository) Host() host.HostDetail {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Host")
	ret0, _ := ret[0].(host.HostDetail)
	return ret0
}

// Host indicates an expected call of Host.
func (mr *MockRepositoryMockRecorder) Host() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Host", reflect.TypeOf((*MockRepository)(nil).Host))
}

// ID mocks base method.
func (m *MockRepository) ID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockRepositoryMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockRepository)(nil).ID))
}

// IsArchived mocks base method.
func (m *MockRepository) IsArchived() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsArchived")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsArchived indicates an expected call of IsArchived.
func (mr *MockRepositoryMockRecorder) IsArchived() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsArchived", reflect.TypeOf((*MockRepository)(nil).IsArchived))
}

// ListPullRequestComments mocks base method.
func (m *MockRepository) ListPullRequestComments(pr *host.PullRequest) ([]host.PullRequestComment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPullRequestComments", pr)
	ret0, _ := ret[0].([]host.PullRequestComment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPullRequestComments indicates an expected call of ListPullRequestComments.
func (mr *MockRepositoryMockRecorder) ListPullRequestComments(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPullRequestComments", reflect.TypeOf((*MockRepository)(nil).ListPullRequestComments), pr)
}

// MergePullRequest mocks base method.
func (m *MockRepository) MergePullRequest(deleteBranch bool, pr *host.PullRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MergePullRequest", deleteBranch, pr)
	ret0, _ := ret[0].(error)
	return ret0
}

// MergePullRequest indicates an expected call of MergePullRequest.
func (mr *MockRepositoryMockRecorder) MergePullRequest(deleteBranch, pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MergePullRequest", reflect.TypeOf((*MockRepository)(nil).MergePullRequest), deleteBranch, pr)
}

// Name mocks base method.
func (m *MockRepository) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockRepositoryMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockRepository)(nil).Name))
}

// Owner mocks base method.
func (m *MockRepository) Owner() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Owner")
	ret0, _ := ret[0].(string)
	return ret0
}

// Owner indicates an expected call of Owner.
func (mr *MockRepositoryMockRecorder) Owner() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Owner", reflect.TypeOf((*MockRepository)(nil).Owner))
}

// Raw mocks base method.
func (m *MockRepository) Raw() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Raw")
	ret0, _ := ret[0].(any)
	return ret0
}

// Raw indicates an expected call of Raw.
func (mr *MockRepositoryMockRecorder) Raw() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Raw", reflect.TypeOf((*MockRepository)(nil).Raw))
}

// UpdatePullRequest mocks base method.
func (m *MockRepository) UpdatePullRequest(data host.PullRequestData, pr *host.PullRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePullRequest", data, pr)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePullRequest indicates an expected call of UpdatePullRequest.
func (mr *MockRepositoryMockRecorder) UpdatePullRequest(data, pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePullRequest", reflect.TypeOf((*MockRepository)(nil).UpdatePullRequest), data, pr)
}

// UpdatedAt mocks base method.
func (m *MockRepository) UpdatedAt() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatedAt")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// UpdatedAt indicates an expected call of UpdatedAt.
func (mr *MockRepositoryMockRecorder) UpdatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatedAt", reflect.TypeOf((*MockRepository)(nil).UpdatedAt))
}

// WebUrl mocks base method.
func (m *MockRepository) WebUrl() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WebUrl")
	ret0, _ := ret[0].(string)
	return ret0
}

// WebUrl indicates an expected call of WebUrl.
func (mr *MockRepositoryMockRecorder) WebUrl() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WebUrl", reflect.TypeOf((*MockRepository)(nil).WebUrl))
}

// MockHost is a mock of Host interface.
type MockHost struct {
	ctrl     *gomock.Controller
	recorder *MockHostMockRecorder
	isgomock struct{}
}

// MockHostMockRecorder is the mock recorder for MockHost.
type MockHostMockRecorder struct {
	mock *MockHost
}

// NewMockHost creates a new mock instance.
func NewMockHost(ctrl *gomock.Controller) *MockHost {
	mock := &MockHost{ctrl: ctrl}
	mock.recorder = &MockHostMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHost) EXPECT() *MockHostMockRecorder {
	return m.recorder
}

// AuthenticatedUser mocks base method.
func (m *MockHost) AuthenticatedUser() (*host.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthenticatedUser")
	ret0, _ := ret[0].(*host.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthenticatedUser indicates an expected call of AuthenticatedUser.
func (mr *MockHostMockRecorder) AuthenticatedUser() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthenticatedUser", reflect.TypeOf((*MockHost)(nil).AuthenticatedUser))
}

// CreateFromJson mocks base method.
func (m *MockHost) CreateFromJson(dec *json.Decoder) (host.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFromJson", dec)
	ret0, _ := ret[0].(host.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFromJson indicates an expected call of CreateFromJson.
func (mr *MockHostMockRecorder) CreateFromJson(dec any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFromJson", reflect.TypeOf((*MockHost)(nil).CreateFromJson), dec)
}

// CreateFromName mocks base method.
func (m *MockHost) CreateFromName(name string) (host.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFromName", name)
	ret0, _ := ret[0].(host.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFromName indicates an expected call of CreateFromName.
func (mr *MockHostMockRecorder) CreateFromName(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFromName", reflect.TypeOf((*MockHost)(nil).CreateFromName), name)
}

// Name mocks base method.
func (m *MockHost) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockHostMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockHost)(nil).Name))
}

// PullRequestFactory mocks base method.
func (m *MockHost) PullRequestFactory() host.PullRequestFactory {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullRequestFactory")
	ret0, _ := ret[0].(host.PullRequestFactory)
	return ret0
}

// PullRequestFactory indicates an expected call of PullRequestFactory.
func (mr *MockHostMockRecorder) PullRequestFactory() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullRequestFactory", reflect.TypeOf((*MockHost)(nil).PullRequestFactory))
}

// PullRequestIterator mocks base method.
func (m *MockHost) PullRequestIterator() host.PullRequestIterator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullRequestIterator")
	ret0, _ := ret[0].(host.PullRequestIterator)
	return ret0
}

// PullRequestIterator indicates an expected call of PullRequestIterator.
func (mr *MockHostMockRecorder) PullRequestIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullRequestIterator", reflect.TypeOf((*MockHost)(nil).PullRequestIterator))
}

// RepositoryIterator mocks base method.
func (m *MockHost) RepositoryIterator() host.RepositoryIterator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RepositoryIterator")
	ret0, _ := ret[0].(host.RepositoryIterator)
	return ret0
}

// RepositoryIterator indicates an expected call of RepositoryIterator.
func (mr *MockHostMockRecorder) RepositoryIterator() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RepositoryIterator", reflect.TypeOf((*MockHost)(nil).RepositoryIterator))
}

// Type mocks base method.
func (m *MockHost) Type() host.Type {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Type")
	ret0, _ := ret[0].(host.Type)
	return ret0
}

// Type indicates an expected call of Type.
func (mr *MockHostMockRecorder) Type() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Type", reflect.TypeOf((*MockHost)(nil).Type))
}

// MockPullRequestIterator is a mock of PullRequestIterator interface.
type MockPullRequestIterator struct {
	ctrl     *gomock.Controller
	recorder *MockPullRequestIteratorMockRecorder
	isgomock struct{}
}

// MockPullRequestIteratorMockRecorder is the mock recorder for MockPullRequestIterator.
type MockPullRequestIteratorMockRecorder struct {
	mock *MockPullRequestIterator
}

// NewMockPullRequestIterator creates a new mock instance.
func NewMockPullRequestIterator(ctrl *gomock.Controller) *MockPullRequestIterator {
	mock := &MockPullRequestIterator{ctrl: ctrl}
	mock.recorder = &MockPullRequestIteratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPullRequestIterator) EXPECT() *MockPullRequestIteratorMockRecorder {
	return m.recorder
}

// Error mocks base method.
func (m *MockPullRequestIterator) Error() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Error")
	ret0, _ := ret[0].(error)
	return ret0
}

// Error indicates an expected call of Error.
func (mr *MockPullRequestIteratorMockRecorder) Error() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockPullRequestIterator)(nil).Error))
}

// ListPullRequests mocks base method.
func (m *MockPullRequestIterator) ListPullRequests(since *time.Time) iter.Seq[*host.PullRequest] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPullRequests", since)
	ret0, _ := ret[0].(iter.Seq[*host.PullRequest])
	return ret0
}

// ListPullRequests indicates an expected call of ListPullRequests.
func (mr *MockPullRequestIteratorMockRecorder) ListPullRequests(since any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPullRequests", reflect.TypeOf((*MockPullRequestIterator)(nil).ListPullRequests), since)
}

// MockRepositoryIterator is a mock of RepositoryIterator interface.
type MockRepositoryIterator struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryIteratorMockRecorder
	isgomock struct{}
}

// MockRepositoryIteratorMockRecorder is the mock recorder for MockRepositoryIterator.
type MockRepositoryIteratorMockRecorder struct {
	mock *MockRepositoryIterator
}

// NewMockRepositoryIterator creates a new mock instance.
func NewMockRepositoryIterator(ctrl *gomock.Controller) *MockRepositoryIterator {
	mock := &MockRepositoryIterator{ctrl: ctrl}
	mock.recorder = &MockRepositoryIteratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoryIterator) EXPECT() *MockRepositoryIteratorMockRecorder {
	return m.recorder
}

// Error mocks base method.
func (m *MockRepositoryIterator) Error() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Error")
	ret0, _ := ret[0].(error)
	return ret0
}

// Error indicates an expected call of Error.
func (mr *MockRepositoryIteratorMockRecorder) Error() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockRepositoryIterator)(nil).Error))
}

// ListRepositories mocks base method.
func (m *MockRepositoryIterator) ListRepositories(since *time.Time) iter.Seq[host.Repository] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRepositories", since)
	ret0, _ := ret[0].(iter.Seq[host.Repository])
	return ret0
}

// ListRepositories indicates an expected call of ListRepositories.
func (mr *MockRepositoryIteratorMockRecorder) ListRepositories(since any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRepositories", reflect.TypeOf((*MockRepositoryIterator)(nil).ListRepositories), since)
}

// MockHostDetail is a mock of HostDetail interface.
type MockHostDetail struct {
	ctrl     *gomock.Controller
	recorder *MockHostDetailMockRecorder
	isgomock struct{}
}

// MockHostDetailMockRecorder is the mock recorder for MockHostDetail.
type MockHostDetailMockRecorder struct {
	mock *MockHostDetail
}

// NewMockHostDetail creates a new mock instance.
func NewMockHostDetail(ctrl *gomock.Controller) *MockHostDetail {
	mock := &MockHostDetail{ctrl: ctrl}
	mock.recorder = &MockHostDetailMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHostDetail) EXPECT() *MockHostDetailMockRecorder {
	return m.recorder
}

// AuthenticatedUser mocks base method.
func (m *MockHostDetail) AuthenticatedUser() (*host.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthenticatedUser")
	ret0, _ := ret[0].(*host.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthenticatedUser indicates an expected call of AuthenticatedUser.
func (mr *MockHostDetailMockRecorder) AuthenticatedUser() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthenticatedUser", reflect.TypeOf((*MockHostDetail)(nil).AuthenticatedUser))
}

// Name mocks base method.
func (m *MockHostDetail) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockHostDetailMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockHostDetail)(nil).Name))
}

// MockRepositoryLister is a mock of RepositoryLister interface.
type MockRepositoryLister struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryListerMockRecorder
	isgomock struct{}
}

// MockRepositoryListerMockRecorder is the mock recorder for MockRepositoryLister.
type MockRepositoryListerMockRecorder struct {
	mock *MockRepositoryLister
}

// NewMockRepositoryLister creates a new mock instance.
func NewMockRepositoryLister(ctrl *gomock.Controller) *MockRepositoryLister {
	mock := &MockRepositoryLister{ctrl: ctrl}
	mock.recorder = &MockRepositoryListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoryLister) EXPECT() *MockRepositoryListerMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockRepositoryLister) List(hosts []host.Host, result chan host.Repository, errChan chan error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "List", hosts, result, errChan)
}

// List indicates an expected call of List.
func (mr *MockRepositoryListerMockRecorder) List(hosts, result, errChan any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRepositoryLister)(nil).List), hosts, result, errChan)
}
