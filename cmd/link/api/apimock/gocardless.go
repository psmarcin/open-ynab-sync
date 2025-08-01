// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package api

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"psmarcin.github.com/open-ynab-sync/cmd/link/api"
)

// NewMockGoCardlessClient creates a new instance of MockGoCardlessClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGoCardlessClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGoCardlessClient {
	mock := &MockGoCardlessClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockGoCardlessClient is an autogenerated mock type for the GoCardlessClient type
type MockGoCardlessClient struct {
	mock.Mock
}

type MockGoCardlessClient_Expecter struct {
	mock *mock.Mock
}

func (_m *MockGoCardlessClient) EXPECT() *MockGoCardlessClient_Expecter {
	return &MockGoCardlessClient_Expecter{mock: &_m.Mock}
}

// CreateAgreement provides a mock function for the type MockGoCardlessClient
func (_mock *MockGoCardlessClient) CreateAgreement(ctx context.Context, institutionID string) (string, error) {
	ret := _mock.Called(ctx, institutionID)

	if len(ret) == 0 {
		panic("no return value specified for CreateAgreement")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return returnFunc(ctx, institutionID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = returnFunc(ctx, institutionID)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, institutionID)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockGoCardlessClient_CreateAgreement_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateAgreement'
type MockGoCardlessClient_CreateAgreement_Call struct {
	*mock.Call
}

// CreateAgreement is a helper method to define mock.On call
//   - ctx context.Context
//   - institutionID string
func (_e *MockGoCardlessClient_Expecter) CreateAgreement(ctx interface{}, institutionID interface{}) *MockGoCardlessClient_CreateAgreement_Call {
	return &MockGoCardlessClient_CreateAgreement_Call{Call: _e.mock.On("CreateAgreement", ctx, institutionID)}
}

func (_c *MockGoCardlessClient_CreateAgreement_Call) Run(run func(ctx context.Context, institutionID string)) *MockGoCardlessClient_CreateAgreement_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockGoCardlessClient_CreateAgreement_Call) Return(s string, err error) *MockGoCardlessClient_CreateAgreement_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockGoCardlessClient_CreateAgreement_Call) RunAndReturn(run func(ctx context.Context, institutionID string) (string, error)) *MockGoCardlessClient_CreateAgreement_Call {
	_c.Call.Return(run)
	return _c
}

// CreateRequisition provides a mock function for the type MockGoCardlessClient
func (_mock *MockGoCardlessClient) CreateRequisition(ctx context.Context, institutionID string, agreementID string, redirectURL string) (string, string, error) {
	ret := _mock.Called(ctx, institutionID, agreementID, redirectURL)

	if len(ret) == 0 {
		panic("no return value specified for CreateRequisition")
	}

	var r0 string
	var r1 string
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string) (string, string, error)); ok {
		return returnFunc(ctx, institutionID, agreementID, redirectURL)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string) string); ok {
		r0 = returnFunc(ctx, institutionID, agreementID, redirectURL)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, string) string); ok {
		r1 = returnFunc(ctx, institutionID, agreementID, redirectURL)
	} else {
		r1 = ret.Get(1).(string)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, string, string) error); ok {
		r2 = returnFunc(ctx, institutionID, agreementID, redirectURL)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockGoCardlessClient_CreateRequisition_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateRequisition'
type MockGoCardlessClient_CreateRequisition_Call struct {
	*mock.Call
}

// CreateRequisition is a helper method to define mock.On call
//   - ctx context.Context
//   - institutionID string
//   - agreementID string
//   - redirectURL string
func (_e *MockGoCardlessClient_Expecter) CreateRequisition(ctx interface{}, institutionID interface{}, agreementID interface{}, redirectURL interface{}) *MockGoCardlessClient_CreateRequisition_Call {
	return &MockGoCardlessClient_CreateRequisition_Call{Call: _e.mock.On("CreateRequisition", ctx, institutionID, agreementID, redirectURL)}
}

func (_c *MockGoCardlessClient_CreateRequisition_Call) Run(run func(ctx context.Context, institutionID string, agreementID string, redirectURL string)) *MockGoCardlessClient_CreateRequisition_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 string
		if args[2] != nil {
			arg2 = args[2].(string)
		}
		var arg3 string
		if args[3] != nil {
			arg3 = args[3].(string)
		}
		run(
			arg0,
			arg1,
			arg2,
			arg3,
		)
	})
	return _c
}

func (_c *MockGoCardlessClient_CreateRequisition_Call) Return(s string, s1 string, err error) *MockGoCardlessClient_CreateRequisition_Call {
	_c.Call.Return(s, s1, err)
	return _c
}

func (_c *MockGoCardlessClient_CreateRequisition_Call) RunAndReturn(run func(ctx context.Context, institutionID string, agreementID string, redirectURL string) (string, string, error)) *MockGoCardlessClient_CreateRequisition_Call {
	_c.Call.Return(run)
	return _c
}

// GetRequisitionStatus provides a mock function for the type MockGoCardlessClient
func (_mock *MockGoCardlessClient) GetRequisitionStatus(ctx context.Context, requisitionID string) (string, []string, error) {
	ret := _mock.Called(ctx, requisitionID)

	if len(ret) == 0 {
		panic("no return value specified for GetRequisitionStatus")
	}

	var r0 string
	var r1 []string
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (string, []string, error)); ok {
		return returnFunc(ctx, requisitionID)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = returnFunc(ctx, requisitionID)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) []string); ok {
		r1 = returnFunc(ctx, requisitionID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]string)
		}
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = returnFunc(ctx, requisitionID)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockGoCardlessClient_GetRequisitionStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRequisitionStatus'
type MockGoCardlessClient_GetRequisitionStatus_Call struct {
	*mock.Call
}

// GetRequisitionStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - requisitionID string
func (_e *MockGoCardlessClient_Expecter) GetRequisitionStatus(ctx interface{}, requisitionID interface{}) *MockGoCardlessClient_GetRequisitionStatus_Call {
	return &MockGoCardlessClient_GetRequisitionStatus_Call{Call: _e.mock.On("GetRequisitionStatus", ctx, requisitionID)}
}

func (_c *MockGoCardlessClient_GetRequisitionStatus_Call) Run(run func(ctx context.Context, requisitionID string)) *MockGoCardlessClient_GetRequisitionStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockGoCardlessClient_GetRequisitionStatus_Call) Return(s string, strings []string, err error) *MockGoCardlessClient_GetRequisitionStatus_Call {
	_c.Call.Return(s, strings, err)
	return _c
}

func (_c *MockGoCardlessClient_GetRequisitionStatus_Call) RunAndReturn(run func(ctx context.Context, requisitionID string) (string, []string, error)) *MockGoCardlessClient_GetRequisitionStatus_Call {
	_c.Call.Return(run)
	return _c
}

// ListRequisitions provides a mock function for the type MockGoCardlessClient
func (_mock *MockGoCardlessClient) ListRequisitions(ctx context.Context) ([]api.Requisition, error) {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for ListRequisitions")
	}

	var r0 []api.Requisition
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) ([]api.Requisition, error)); ok {
		return returnFunc(ctx)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context) []api.Requisition); ok {
		r0 = returnFunc(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]api.Requisition)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = returnFunc(ctx)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockGoCardlessClient_ListRequisitions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListRequisitions'
type MockGoCardlessClient_ListRequisitions_Call struct {
	*mock.Call
}

// ListRequisitions is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockGoCardlessClient_Expecter) ListRequisitions(ctx interface{}) *MockGoCardlessClient_ListRequisitions_Call {
	return &MockGoCardlessClient_ListRequisitions_Call{Call: _e.mock.On("ListRequisitions", ctx)}
}

func (_c *MockGoCardlessClient_ListRequisitions_Call) Run(run func(ctx context.Context)) *MockGoCardlessClient_ListRequisitions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockGoCardlessClient_ListRequisitions_Call) Return(requisitions []api.Requisition, err error) *MockGoCardlessClient_ListRequisitions_Call {
	_c.Call.Return(requisitions, err)
	return _c
}

func (_c *MockGoCardlessClient_ListRequisitions_Call) RunAndReturn(run func(ctx context.Context) ([]api.Requisition, error)) *MockGoCardlessClient_ListRequisitions_Call {
	_c.Call.Return(run)
	return _c
}

// LogIn provides a mock function for the type MockGoCardlessClient
func (_mock *MockGoCardlessClient) LogIn(ctx context.Context) error {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for LogIn")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = returnFunc(ctx)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockGoCardlessClient_LogIn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LogIn'
type MockGoCardlessClient_LogIn_Call struct {
	*mock.Call
}

// LogIn is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockGoCardlessClient_Expecter) LogIn(ctx interface{}) *MockGoCardlessClient_LogIn_Call {
	return &MockGoCardlessClient_LogIn_Call{Call: _e.mock.On("LogIn", ctx)}
}

func (_c *MockGoCardlessClient_LogIn_Call) Run(run func(ctx context.Context)) *MockGoCardlessClient_LogIn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockGoCardlessClient_LogIn_Call) Return(err error) *MockGoCardlessClient_LogIn_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockGoCardlessClient_LogIn_Call) RunAndReturn(run func(ctx context.Context) error) *MockGoCardlessClient_LogIn_Call {
	_c.Call.Return(run)
	return _c
}
