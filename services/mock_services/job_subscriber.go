// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/smartcontractkit/chainlink/services (interfaces: JobSubscriber)

// Package mock_services is a generated GoMock package.
package mock_services

import (
	gomock "github.com/golang/mock/gomock"
	models "github.com/smartcontractkit/chainlink/store/models"
	reflect "reflect"
)

// MockJobSubscriber is a mock of JobSubscriber interface
type MockJobSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockJobSubscriberMockRecorder
}

// MockJobSubscriberMockRecorder is the mock recorder for MockJobSubscriber
type MockJobSubscriberMockRecorder struct {
	mock *MockJobSubscriber
}

// NewMockJobSubscriber creates a new mock instance
func NewMockJobSubscriber(ctrl *gomock.Controller) *MockJobSubscriber {
	mock := &MockJobSubscriber{ctrl: ctrl}
	mock.recorder = &MockJobSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockJobSubscriber) EXPECT() *MockJobSubscriberMockRecorder {
	return m.recorder
}

// AddJob mocks base method
func (m *MockJobSubscriber) AddJob(arg0 models.JobSpec, arg1 *models.IndexableBlockNumber) error {
	ret := m.ctrl.Call(m, "AddJob", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddJob indicates an expected call of AddJob
func (mr *MockJobSubscriberMockRecorder) AddJob(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddJob", reflect.TypeOf((*MockJobSubscriber)(nil).AddJob), arg0, arg1)
}

// Connect mocks base method
func (m *MockJobSubscriber) Connect(arg0 *models.IndexableBlockNumber) error {
	ret := m.ctrl.Call(m, "Connect", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect
func (mr *MockJobSubscriberMockRecorder) Connect(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockJobSubscriber)(nil).Connect), arg0)
}

// Disconnect mocks base method
func (m *MockJobSubscriber) Disconnect() {
	m.ctrl.Call(m, "Disconnect")
}

// Disconnect indicates an expected call of Disconnect
func (mr *MockJobSubscriberMockRecorder) Disconnect() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockJobSubscriber)(nil).Disconnect))
}

// Jobs mocks base method
func (m *MockJobSubscriber) Jobs() []models.JobSpec {
	ret := m.ctrl.Call(m, "Jobs")
	ret0, _ := ret[0].([]models.JobSpec)
	return ret0
}

// Jobs indicates an expected call of Jobs
func (mr *MockJobSubscriberMockRecorder) Jobs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Jobs", reflect.TypeOf((*MockJobSubscriber)(nil).Jobs))
}

// OnNewHead mocks base method
func (m *MockJobSubscriber) OnNewHead(arg0 *models.BlockHeader) {
	m.ctrl.Call(m, "OnNewHead", arg0)
}

// OnNewHead indicates an expected call of OnNewHead
func (mr *MockJobSubscriberMockRecorder) OnNewHead(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnNewHead", reflect.TypeOf((*MockJobSubscriber)(nil).OnNewHead), arg0)
}
