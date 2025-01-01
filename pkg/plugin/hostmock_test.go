// Code generated by MockGen. DO NOT EDIT.
// Source: ../host/host.go
//
// Generated by this command:
//
//	mockgen -package plugin_test -source ../host/host.go -destination hostmock_test.go -write_generate_directive
//

// Package plugin_test is a generated GoMock package.
package plugin_test

import (
	json "encoding/json"
	reflect "reflect"
	time "time"

	host "github.com/wndhydrnt/saturn-bot/pkg/host"
	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -package plugin_test -source ../host/host.go -destination hostmock_test.go -write_generate_directive

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
func (m *MockRepository) CanMergePullRequest(pr any) (bool, error) {
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
func (m *MockRepository) ClosePullRequest(msg string, pr any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ClosePullRequest", msg, pr)
	ret0, _ := ret[0].(error)
	return ret0
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
func (m *MockRepository) CreatePullRequestComment(body string, pr any) error {
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
func (m *MockRepository) DeleteBranch(pr any) error {
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
func (m *MockRepository) DeletePullRequestComment(comment host.PullRequestComment, pr any) error {
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
func (m *MockRepository) FindPullRequest(branch string) (any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindPullRequest", branch)
	ret0, _ := ret[0].(any)
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
func (m *MockRepository) GetPullRequestBody(pr any) string {
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

// GetPullRequestCreationTime mocks base method.
func (m *MockRepository) GetPullRequestCreationTime(pr any) time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPullRequestCreationTime", pr)
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetPullRequestCreationTime indicates an expected call of GetPullRequestCreationTime.
func (mr *MockRepositoryMockRecorder) GetPullRequestCreationTime(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPullRequestCreationTime", reflect.TypeOf((*MockRepository)(nil).GetPullRequestCreationTime), pr)
}

// HasSuccessfulPullRequestBuild mocks base method.
func (m *MockRepository) HasSuccessfulPullRequestBuild(pr any) (bool, error) {
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

// IsPullRequestClosed mocks base method.
func (m *MockRepository) IsPullRequestClosed(pr any) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPullRequestClosed", pr)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsPullRequestClosed indicates an expected call of IsPullRequestClosed.
func (mr *MockRepositoryMockRecorder) IsPullRequestClosed(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPullRequestClosed", reflect.TypeOf((*MockRepository)(nil).IsPullRequestClosed), pr)
}

// IsPullRequestMerged mocks base method.
func (m *MockRepository) IsPullRequestMerged(pr any) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPullRequestMerged", pr)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsPullRequestMerged indicates an expected call of IsPullRequestMerged.
func (mr *MockRepositoryMockRecorder) IsPullRequestMerged(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPullRequestMerged", reflect.TypeOf((*MockRepository)(nil).IsPullRequestMerged), pr)
}

// IsPullRequestOpen mocks base method.
func (m *MockRepository) IsPullRequestOpen(pr any) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsPullRequestOpen", pr)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsPullRequestOpen indicates an expected call of IsPullRequestOpen.
func (mr *MockRepositoryMockRecorder) IsPullRequestOpen(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsPullRequestOpen", reflect.TypeOf((*MockRepository)(nil).IsPullRequestOpen), pr)
}

// ListPullRequestComments mocks base method.
func (m *MockRepository) ListPullRequestComments(pr any) ([]host.PullRequestComment, error) {
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
func (m *MockRepository) MergePullRequest(deleteBranch bool, pr any) error {
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

// PullRequest mocks base method.
func (m *MockRepository) PullRequest(pr any) *host.PullRequest {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullRequest", pr)
	ret0, _ := ret[0].(*host.PullRequest)
	return ret0
}

// PullRequest indicates an expected call of PullRequest.
func (mr *MockRepositoryMockRecorder) PullRequest(pr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullRequest", reflect.TypeOf((*MockRepository)(nil).PullRequest), pr)
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
func (m *MockRepository) UpdatePullRequest(data host.PullRequestData, pr any) error {
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

// ListRepositories mocks base method.
func (m *MockHost) ListRepositories(since *time.Time, result chan []host.Repository, errChan chan error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ListRepositories", since, result, errChan)
}

// ListRepositories indicates an expected call of ListRepositories.
func (mr *MockHostMockRecorder) ListRepositories(since, result, errChan any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRepositories", reflect.TypeOf((*MockHost)(nil).ListRepositories), since, result, errChan)
}

// ListRepositoriesWithOpenPullRequests mocks base method.
func (m *MockHost) ListRepositoriesWithOpenPullRequests(result chan []host.Repository, errChan chan error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ListRepositoriesWithOpenPullRequests", result, errChan)
}

// ListRepositoriesWithOpenPullRequests indicates an expected call of ListRepositoriesWithOpenPullRequests.
func (mr *MockHostMockRecorder) ListRepositoriesWithOpenPullRequests(result, errChan any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRepositoriesWithOpenPullRequests", reflect.TypeOf((*MockHost)(nil).ListRepositoriesWithOpenPullRequests), result, errChan)
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

// MockRepositoryCacheLister is a mock of RepositoryCacheLister interface.
type MockRepositoryCacheLister struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryCacheListerMockRecorder
	isgomock struct{}
}

// MockRepositoryCacheListerMockRecorder is the mock recorder for MockRepositoryCacheLister.
type MockRepositoryCacheListerMockRecorder struct {
	mock *MockRepositoryCacheLister
}

// NewMockRepositoryCacheLister creates a new mock instance.
func NewMockRepositoryCacheLister(ctrl *gomock.Controller) *MockRepositoryCacheLister {
	mock := &MockRepositoryCacheLister{ctrl: ctrl}
	mock.recorder = &MockRepositoryCacheListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoryCacheLister) EXPECT() *MockRepositoryCacheListerMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockRepositoryCacheLister) List(hosts []host.Host, result chan host.Repository, errChan chan error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "List", hosts, result, errChan)
}

// List indicates an expected call of List.
func (mr *MockRepositoryCacheListerMockRecorder) List(hosts, result, errChan any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRepositoryCacheLister)(nil).List), hosts, result, errChan)
}

// MockRepositoryCacheRemover is a mock of RepositoryCacheRemover interface.
type MockRepositoryCacheRemover struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryCacheRemoverMockRecorder
	isgomock struct{}
}

// MockRepositoryCacheRemoverMockRecorder is the mock recorder for MockRepositoryCacheRemover.
type MockRepositoryCacheRemoverMockRecorder struct {
	mock *MockRepositoryCacheRemover
}

// NewMockRepositoryCacheRemover creates a new mock instance.
func NewMockRepositoryCacheRemover(ctrl *gomock.Controller) *MockRepositoryCacheRemover {
	mock := &MockRepositoryCacheRemover{ctrl: ctrl}
	mock.recorder = &MockRepositoryCacheRemoverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepositoryCacheRemover) EXPECT() *MockRepositoryCacheRemoverMockRecorder {
	return m.recorder
}

// Remove mocks base method.
func (m *MockRepositoryCacheRemover) Remove(repo host.Repository) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", repo)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockRepositoryCacheRemoverMockRecorder) Remove(repo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRepositoryCacheRemover)(nil).Remove), repo)
}
