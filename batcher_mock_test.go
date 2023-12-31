// Code generated by MockGen. DO NOT EDIT.
// Source: batcher.go
//
// Generated by this command:
//
//	mockgen -destination=batcher_mock_test.go -source=batcher.go -package=batcher github.com/yckao/go-batcher Batch,Action
//
// Package batcher is a generated GoMock package.
package batcher

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockBatcher is a mock of Batcher interface.
type MockBatcher[REQ any, RES any] struct {
	ctrl     *gomock.Controller
	recorder *MockBatcherMockRecorder[REQ, RES]
}

// MockBatcherMockRecorder is the mock recorder for MockBatcher.
type MockBatcherMockRecorder[REQ any, RES any] struct {
	mock *MockBatcher[REQ, RES]
}

// NewMockBatcher creates a new mock instance.
func NewMockBatcher[REQ any, RES any](ctrl *gomock.Controller) *MockBatcher[REQ, RES] {
	mock := &MockBatcher[REQ, RES]{ctrl: ctrl}
	mock.recorder = &MockBatcherMockRecorder[REQ, RES]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatcher[REQ, RES]) EXPECT() *MockBatcherMockRecorder[REQ, RES] {
	return m.recorder
}

// Do mocks base method.
func (m *MockBatcher[REQ, RES]) Do(arg0 context.Context, arg1 REQ) Thunk[RES] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", arg0, arg1)
	ret0, _ := ret[0].(Thunk[RES])
	return ret0
}

// Do indicates an expected call of Do.
func (mr *MockBatcherMockRecorder[REQ, RES]) Do(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockBatcher[REQ, RES])(nil).Do), arg0, arg1)
}

// Shutdown mocks base method.
func (m *MockBatcher[REQ, RES]) Shutdown() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Shutdown")
	ret0, _ := ret[0].(error)
	return ret0
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockBatcherMockRecorder[REQ, RES]) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockBatcher[REQ, RES])(nil).Shutdown))
}

// MockBatch is a mock of Batch interface.
type MockBatch struct {
	ctrl     *gomock.Controller
	recorder *MockBatchMockRecorder
}

// MockBatchMockRecorder is the mock recorder for MockBatch.
type MockBatchMockRecorder struct {
	mock *MockBatch
}

// NewMockBatch creates a new mock instance.
func NewMockBatch(ctrl *gomock.Controller) *MockBatch {
	mock := &MockBatch{ctrl: ctrl}
	mock.recorder = &MockBatchMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatch) EXPECT() *MockBatchMockRecorder {
	return m.recorder
}

// Dispatch mocks base method.
func (m *MockBatch) Dispatch() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dispatch")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Dispatch indicates an expected call of Dispatch.
func (mr *MockBatchMockRecorder) Dispatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dispatch", reflect.TypeOf((*MockBatch)(nil).Dispatch))
}

// Full mocks base method.
func (m *MockBatch) Full() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Full")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Full indicates an expected call of Full.
func (mr *MockBatchMockRecorder) Full() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Full", reflect.TypeOf((*MockBatch)(nil).Full))
}

// MockAction is a mock of Action interface.
type MockAction[REQ any, RES any] struct {
	ctrl     *gomock.Controller
	recorder *MockActionMockRecorder[REQ, RES]
}

// MockActionMockRecorder is the mock recorder for MockAction.
type MockActionMockRecorder[REQ any, RES any] struct {
	mock *MockAction[REQ, RES]
}

// NewMockAction creates a new mock instance.
func NewMockAction[REQ any, RES any](ctrl *gomock.Controller) *MockAction[REQ, RES] {
	mock := &MockAction[REQ, RES]{ctrl: ctrl}
	mock.recorder = &MockActionMockRecorder[REQ, RES]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAction[REQ, RES]) EXPECT() *MockActionMockRecorder[REQ, RES] {
	return m.recorder
}

// Perform mocks base method.
func (m *MockAction[REQ, RES]) Perform(arg0 context.Context, arg1 []REQ) []Response[RES] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Perform", arg0, arg1)
	ret0, _ := ret[0].([]Response[RES])
	return ret0
}

// Perform indicates an expected call of Perform.
func (mr *MockActionMockRecorder[REQ, RES]) Perform(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Perform", reflect.TypeOf((*MockAction[REQ, RES])(nil).Perform), arg0, arg1)
}
