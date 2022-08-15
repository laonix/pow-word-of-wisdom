// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	net "net"

	mock "github.com/stretchr/testify/mock"
)

// Conn is an autogenerated mock type for the Conn type
type Conn struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Conn) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Read provides a mock function with given fields: b
func (_m *Conn) Read(b []byte) ([]byte, error) {
	ret := _m.Called(b)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]byte) []byte); ok {
		r0 = rf(b)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(b)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoteAddr provides a mock function with given fields:
func (_m *Conn) RemoteAddr() net.Addr {
	ret := _m.Called()

	var r0 net.Addr
	if rf, ok := ret.Get(0).(func() net.Addr); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Addr)
		}
	}

	return r0
}

// Write provides a mock function with given fields: b
func (_m *Conn) Write(b []byte) (int, error) {
	ret := _m.Called(b)

	var r0 int
	if rf, ok := ret.Get(0).(func([]byte) int); ok {
		r0 = rf(b)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(b)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewConn interface {
	mock.TestingT
	Cleanup(func())
}

// NewConn creates a new instance of Conn. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConn(t mockConstructorTestingTNewConn) *Conn {
	mock := &Conn{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
