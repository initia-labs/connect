// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	types "github.com/cometbft/cometbft/abci/types"
	mock "github.com/stretchr/testify/mock"
)

// ExtendedCommitCodec is an autogenerated mock type for the ExtendedCommitCodec type
type ExtendedCommitCodec struct {
	mock.Mock
}

// Decode provides a mock function with given fields: _a0
func (_m *ExtendedCommitCodec) Decode(_a0 []byte) (types.ExtendedCommitInfo, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Decode")
	}

	var r0 types.ExtendedCommitInfo
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (types.ExtendedCommitInfo, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func([]byte) types.ExtendedCommitInfo); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(types.ExtendedCommitInfo)
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Encode provides a mock function with given fields: _a0
func (_m *ExtendedCommitCodec) Encode(_a0 types.ExtendedCommitInfo) ([]byte, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Encode")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(types.ExtendedCommitInfo) ([]byte, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(types.ExtendedCommitInfo) []byte); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(types.ExtendedCommitInfo) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewExtendedCommitCodec creates a new instance of ExtendedCommitCodec. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExtendedCommitCodec(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExtendedCommitCodec {
	mock := &ExtendedCommitCodec{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
