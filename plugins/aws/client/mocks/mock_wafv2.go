// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudquery/cq-provider-aws/client (interfaces: WafV2Client)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	wafv2 "github.com/aws/aws-sdk-go-v2/service/wafv2"
	gomock "github.com/golang/mock/gomock"
)

// MockWafV2Client is a mock of WafV2Client interface.
type MockWafV2Client struct {
	ctrl     *gomock.Controller
	recorder *MockWafV2ClientMockRecorder
}

// MockWafV2ClientMockRecorder is the mock recorder for MockWafV2Client.
type MockWafV2ClientMockRecorder struct {
	mock *MockWafV2Client
}

// NewMockWafV2Client creates a new mock instance.
func NewMockWafV2Client(ctrl *gomock.Controller) *MockWafV2Client {
	mock := &MockWafV2Client{ctrl: ctrl}
	mock.recorder = &MockWafV2ClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWafV2Client) EXPECT() *MockWafV2ClientMockRecorder {
	return m.recorder
}

// DescribeManagedRuleGroup mocks base method.
func (m *MockWafV2Client) DescribeManagedRuleGroup(arg0 context.Context, arg1 *wafv2.DescribeManagedRuleGroupInput, arg2 ...func(*wafv2.Options)) (*wafv2.DescribeManagedRuleGroupOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DescribeManagedRuleGroup", varargs...)
	ret0, _ := ret[0].(*wafv2.DescribeManagedRuleGroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeManagedRuleGroup indicates an expected call of DescribeManagedRuleGroup.
func (mr *MockWafV2ClientMockRecorder) DescribeManagedRuleGroup(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeManagedRuleGroup", reflect.TypeOf((*MockWafV2Client)(nil).DescribeManagedRuleGroup), varargs...)
}

// GetPermissionPolicy mocks base method.
func (m *MockWafV2Client) GetPermissionPolicy(arg0 context.Context, arg1 *wafv2.GetPermissionPolicyInput, arg2 ...func(*wafv2.Options)) (*wafv2.GetPermissionPolicyOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPermissionPolicy", varargs...)
	ret0, _ := ret[0].(*wafv2.GetPermissionPolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPermissionPolicy indicates an expected call of GetPermissionPolicy.
func (mr *MockWafV2ClientMockRecorder) GetPermissionPolicy(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPermissionPolicy", reflect.TypeOf((*MockWafV2Client)(nil).GetPermissionPolicy), varargs...)
}

// GetRuleGroup mocks base method.
func (m *MockWafV2Client) GetRuleGroup(arg0 context.Context, arg1 *wafv2.GetRuleGroupInput, arg2 ...func(*wafv2.Options)) (*wafv2.GetRuleGroupOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRuleGroup", varargs...)
	ret0, _ := ret[0].(*wafv2.GetRuleGroupOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRuleGroup indicates an expected call of GetRuleGroup.
func (mr *MockWafV2ClientMockRecorder) GetRuleGroup(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRuleGroup", reflect.TypeOf((*MockWafV2Client)(nil).GetRuleGroup), varargs...)
}

// GetWebACL mocks base method.
func (m *MockWafV2Client) GetWebACL(arg0 context.Context, arg1 *wafv2.GetWebACLInput, arg2 ...func(*wafv2.Options)) (*wafv2.GetWebACLOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetWebACL", varargs...)
	ret0, _ := ret[0].(*wafv2.GetWebACLOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWebACL indicates an expected call of GetWebACL.
func (mr *MockWafV2ClientMockRecorder) GetWebACL(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWebACL", reflect.TypeOf((*MockWafV2Client)(nil).GetWebACL), varargs...)
}

// GetWebACLForResource mocks base method.
func (m *MockWafV2Client) GetWebACLForResource(arg0 context.Context, arg1 *wafv2.GetWebACLForResourceInput, arg2 ...func(*wafv2.Options)) (*wafv2.GetWebACLForResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetWebACLForResource", varargs...)
	ret0, _ := ret[0].(*wafv2.GetWebACLForResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWebACLForResource indicates an expected call of GetWebACLForResource.
func (mr *MockWafV2ClientMockRecorder) GetWebACLForResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWebACLForResource", reflect.TypeOf((*MockWafV2Client)(nil).GetWebACLForResource), varargs...)
}

// ListAvailableManagedRuleGroups mocks base method.
func (m *MockWafV2Client) ListAvailableManagedRuleGroups(arg0 context.Context, arg1 *wafv2.ListAvailableManagedRuleGroupsInput, arg2 ...func(*wafv2.Options)) (*wafv2.ListAvailableManagedRuleGroupsOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListAvailableManagedRuleGroups", varargs...)
	ret0, _ := ret[0].(*wafv2.ListAvailableManagedRuleGroupsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAvailableManagedRuleGroups indicates an expected call of ListAvailableManagedRuleGroups.
func (mr *MockWafV2ClientMockRecorder) ListAvailableManagedRuleGroups(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAvailableManagedRuleGroups", reflect.TypeOf((*MockWafV2Client)(nil).ListAvailableManagedRuleGroups), varargs...)
}

// ListResourcesForWebACL mocks base method.
func (m *MockWafV2Client) ListResourcesForWebACL(arg0 context.Context, arg1 *wafv2.ListResourcesForWebACLInput, arg2 ...func(*wafv2.Options)) (*wafv2.ListResourcesForWebACLOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListResourcesForWebACL", varargs...)
	ret0, _ := ret[0].(*wafv2.ListResourcesForWebACLOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListResourcesForWebACL indicates an expected call of ListResourcesForWebACL.
func (mr *MockWafV2ClientMockRecorder) ListResourcesForWebACL(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListResourcesForWebACL", reflect.TypeOf((*MockWafV2Client)(nil).ListResourcesForWebACL), varargs...)
}

// ListRuleGroups mocks base method.
func (m *MockWafV2Client) ListRuleGroups(arg0 context.Context, arg1 *wafv2.ListRuleGroupsInput, arg2 ...func(*wafv2.Options)) (*wafv2.ListRuleGroupsOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListRuleGroups", varargs...)
	ret0, _ := ret[0].(*wafv2.ListRuleGroupsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRuleGroups indicates an expected call of ListRuleGroups.
func (mr *MockWafV2ClientMockRecorder) ListRuleGroups(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRuleGroups", reflect.TypeOf((*MockWafV2Client)(nil).ListRuleGroups), varargs...)
}

// ListTagsForResource mocks base method.
func (m *MockWafV2Client) ListTagsForResource(arg0 context.Context, arg1 *wafv2.ListTagsForResourceInput, arg2 ...func(*wafv2.Options)) (*wafv2.ListTagsForResourceOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListTagsForResource", varargs...)
	ret0, _ := ret[0].(*wafv2.ListTagsForResourceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTagsForResource indicates an expected call of ListTagsForResource.
func (mr *MockWafV2ClientMockRecorder) ListTagsForResource(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTagsForResource", reflect.TypeOf((*MockWafV2Client)(nil).ListTagsForResource), varargs...)
}

// ListWebACLs mocks base method.
func (m *MockWafV2Client) ListWebACLs(arg0 context.Context, arg1 *wafv2.ListWebACLsInput, arg2 ...func(*wafv2.Options)) (*wafv2.ListWebACLsOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListWebACLs", varargs...)
	ret0, _ := ret[0].(*wafv2.ListWebACLsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListWebACLs indicates an expected call of ListWebACLs.
func (mr *MockWafV2ClientMockRecorder) ListWebACLs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListWebACLs", reflect.TypeOf((*MockWafV2Client)(nil).ListWebACLs), varargs...)
}
