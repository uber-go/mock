// Code generated by MockGen. DO NOT EDIT.
// Source: go.uber.org/mock/mockgen/internal/tests/add_generate_directive (interfaces: Foo)
//
// Generated by this command:
//
//	mockgen -write_generate_directive -destination mock.go -package add_generate_directive . Foo
//

// Package add_generate_directive is a generated GoMock package.
package add_generate_directive

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -write_generate_directive -destination mock.go -package add_generate_directive . Foo

// MockFoo is a mock of Foo interface.
type MockFoo struct {
	ctrl     *gomock.Controller
	recorder *MockFooMockRecorder
}

// MockFooMockRecorder is the mock recorder for MockFoo.
type MockFooMockRecorder struct {
	mock *MockFoo
}

// NewMockFoo creates a new mock instance.
func NewMockFoo(ctrl *gomock.Controller) *MockFoo {
	mock := &MockFoo{ctrl: ctrl}
	mock.recorder = &MockFooMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFoo) EXPECT() *MockFooMockRecorder {
	return m.recorder
}

// ISGOMOCK indicates that this struct is a gomock mock.
func (m *MockFoo) ISGOMOCK() struct{} {
	return struct{}{}
}

// Bar mocks base method.
func (m *MockFoo) Bar(arg0 []string, arg1 chan<- Message) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Bar", arg0, arg1)
}

// Bar indicates an expected call of Bar.
func (mr *MockFooMockRecorder) Bar(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bar", reflect.TypeOf((*MockFoo)(nil).Bar), arg0, arg1)
}
