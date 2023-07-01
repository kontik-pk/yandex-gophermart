// Code generated by mockery. DO NOT EDIT.

package handlers

import mock "github.com/stretchr/testify/mock"

// mockDbManager is an autogenerated mock type for the dbManager type
type mockDbManager struct {
	mock.Mock
}

// GetBalanceInfo provides a mock function with given fields: login
func (_m *mockDbManager) GetBalanceInfo(login string) ([]byte, error) {
	ret := _m.Called(login)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, error)); ok {
		return rf(login)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserOrders provides a mock function with given fields: login
func (_m *mockDbManager) GetUserOrders(login string) ([]byte, error) {
	ret := _m.Called(login)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, error)); ok {
		return rf(login)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWithdrawals provides a mock function with given fields: login
func (_m *mockDbManager) GetWithdrawals(login string) ([]byte, error) {
	ret := _m.Called(login)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, error)); ok {
		return rf(login)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(login)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(login)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadOrder provides a mock function with given fields: login, orderID
func (_m *mockDbManager) LoadOrder(login string, orderID string) error {
	ret := _m.Called(login, orderID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(login, orderID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Login provides a mock function with given fields: login, password
func (_m *mockDbManager) Login(login string, password string) error {
	ret := _m.Called(login, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(login, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Register provides a mock function with given fields: login, password
func (_m *mockDbManager) Register(login string, password string) error {
	ret := _m.Called(login, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(login, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Withdraw provides a mock function with given fields: login, orderID, sum
func (_m *mockDbManager) Withdraw(login string, orderID string, sum float64) error {
	ret := _m.Called(login, orderID, sum)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, float64) error); ok {
		r0 = rf(login, orderID, sum)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTnewMockDbManager interface {
	mock.TestingT
	Cleanup(func())
}

// newMockDbManager creates a new instance of mockDbManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newMockDbManager(t mockConstructorTestingTnewMockDbManager) *mockDbManager {
	mock := &mockDbManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}