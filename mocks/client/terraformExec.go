// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	context "context"

	tfexec "github.com/hashicorp/terraform-exec/tfexec"
	mock "github.com/stretchr/testify/mock"
)

// TerraformExec is an autogenerated mock type for the terraformExec type
type TerraformExec struct {
	mock.Mock
}

// Apply provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.ApplyOption) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Init provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.InitOption) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Plan provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.PlanOption) bool); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...tfexec.PlanOption) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetEnv provides a mock function with given fields: env
func (_m *TerraformExec) SetEnv(env map[string]string) error {
	ret := _m.Called(env)

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]string) error); ok {
		r0 = rf(env)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WorkspaceNew provides a mock function with given fields: ctx, workspace, opts
func (_m *TerraformExec) WorkspaceNew(ctx context.Context, workspace string, opts ...tfexec.WorkspaceNewCmdOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, workspace)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...tfexec.WorkspaceNewCmdOption) error); ok {
		r0 = rf(ctx, workspace, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WorkspaceSelect provides a mock function with given fields: ctx, workspace
func (_m *TerraformExec) WorkspaceSelect(ctx context.Context, workspace string) error {
	ret := _m.Called(ctx, workspace)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, workspace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
