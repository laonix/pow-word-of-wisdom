// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// WordOfWisdom is an autogenerated mock type for the WordOfWisdom type
type WordOfWisdom struct {
	mock.Mock
}

// Quote provides a mock function with given fields:
func (_m *WordOfWisdom) Quote() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewWordOfWisdom interface {
	mock.TestingT
	Cleanup(func())
}

// NewWordOfWisdom creates a new instance of WordOfWisdom. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWordOfWisdom(t mockConstructorTestingTNewWordOfWisdom) *WordOfWisdom {
	mock := &WordOfWisdom{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
