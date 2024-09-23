// Code generated by MockGen. DO NOT EDIT.
// Source: input.go
//
// Generated by this command:
//
//	mockgen -package defined_import_local_name -destination mock.go -source input.go -imports b_mock=bytes,c_mock=context
//

// Package defined_import_local_name is a generated GoMock package.
package defined_import_local_name

import (
	b_mock "bytes"
	c_mock "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockWithImports is a mock of WithImports interface.
type MockWithImports struct {
	ctrl     *gomock.Controller
	recorder *MockWithImportsMockRecorder
	isgomock struct{}
}

// MockWithImportsMockRecorder is the mock recorder for MockWithImports.
type MockWithImportsMockRecorder struct {
	mock *MockWithImports
}

// NewMockWithImports creates a new mock instance.
func NewMockWithImports(ctrl *gomock.Controller) *MockWithImports {
	mock := &MockWithImports{ctrl: ctrl}
	mock.recorder = &MockWithImportsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWithImports) EXPECT() *MockWithImportsMockRecorder {
	return m.recorder
}

// Method1 mocks base method.
func (m *MockWithImports) Method1() b_mock.Buffer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Method1")
	ret0, _ := ret[0].(b_mock.Buffer)
	return ret0
}

// Method1 indicates an expected call of Method1.
func (mr *MockWithImportsMockRecorder) Method1() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Method1", reflect.TypeOf((*MockWithImports)(nil).Method1))
}

// Method2 mocks base method.
func (m *MockWithImports) Method2() c_mock.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Method2")
	ret0, _ := ret[0].(c_mock.Context)
	return ret0
}

// Method2 indicates an expected call of Method2.
func (mr *MockWithImportsMockRecorder) Method2() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Method2", reflect.TypeOf((*MockWithImports)(nil).Method2))
}
