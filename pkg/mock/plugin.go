// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/wndhydrnt/saturn-bot-go/plugin (interfaces: Provider)
//
// Generated by this command:
//
//	mockgen -package mock github.com/wndhydrnt/saturn-bot-go/plugin Provider
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	protocolv1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	gomock "go.uber.org/mock/gomock"
)

// MockProvider is a mock of Provider interface.
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
}

// MockProviderMockRecorder is the mock recorder for MockProvider.
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance.
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// ExecuteActions mocks base method.
func (m *MockProvider) ExecuteActions(arg0 *protocolv1.ExecuteActionsRequest) (*protocolv1.ExecuteActionsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteActions", arg0)
	ret0, _ := ret[0].(*protocolv1.ExecuteActionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecuteActions indicates an expected call of ExecuteActions.
func (mr *MockProviderMockRecorder) ExecuteActions(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteActions", reflect.TypeOf((*MockProvider)(nil).ExecuteActions), arg0)
}

// ExecuteFilters mocks base method.
func (m *MockProvider) ExecuteFilters(arg0 *protocolv1.ExecuteFiltersRequest) (*protocolv1.ExecuteFiltersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteFilters", arg0)
	ret0, _ := ret[0].(*protocolv1.ExecuteFiltersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecuteFilters indicates an expected call of ExecuteFilters.
func (mr *MockProviderMockRecorder) ExecuteFilters(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteFilters", reflect.TypeOf((*MockProvider)(nil).ExecuteFilters), arg0)
}

// GetPlugin mocks base method.
func (m *MockProvider) GetPlugin(arg0 *protocolv1.GetPluginRequest) (*protocolv1.GetPluginResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlugin", arg0)
	ret0, _ := ret[0].(*protocolv1.GetPluginResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPlugin indicates an expected call of GetPlugin.
func (mr *MockProviderMockRecorder) GetPlugin(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlugin", reflect.TypeOf((*MockProvider)(nil).GetPlugin), arg0)
}

// OnPrClosed mocks base method.
func (m *MockProvider) OnPrClosed(arg0 *protocolv1.OnPrClosedRequest) (*protocolv1.OnPrClosedResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnPrClosed", arg0)
	ret0, _ := ret[0].(*protocolv1.OnPrClosedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OnPrClosed indicates an expected call of OnPrClosed.
func (mr *MockProviderMockRecorder) OnPrClosed(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPrClosed", reflect.TypeOf((*MockProvider)(nil).OnPrClosed), arg0)
}

// OnPrCreated mocks base method.
func (m *MockProvider) OnPrCreated(arg0 *protocolv1.OnPrCreatedRequest) (*protocolv1.OnPrCreatedResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnPrCreated", arg0)
	ret0, _ := ret[0].(*protocolv1.OnPrCreatedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OnPrCreated indicates an expected call of OnPrCreated.
func (mr *MockProviderMockRecorder) OnPrCreated(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPrCreated", reflect.TypeOf((*MockProvider)(nil).OnPrCreated), arg0)
}

// OnPrMerged mocks base method.
func (m *MockProvider) OnPrMerged(arg0 *protocolv1.OnPrMergedRequest) (*protocolv1.OnPrMergedResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OnPrMerged", arg0)
	ret0, _ := ret[0].(*protocolv1.OnPrMergedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OnPrMerged indicates an expected call of OnPrMerged.
func (mr *MockProviderMockRecorder) OnPrMerged(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnPrMerged", reflect.TypeOf((*MockProvider)(nil).OnPrMerged), arg0)
}