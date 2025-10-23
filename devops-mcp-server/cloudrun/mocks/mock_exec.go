// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law of an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package mocks

import (
	"os/exec"
	"reflect"

	"github.com/golang/mock/gomock"
)

// MockExec is a mock of Exec interface.
type MockExec struct {
	ctrl     *gomock.Controller
	recorder *MockExecMockRecorder
}

// MockExecMockRecorder is the mock recorder for MockExec.
type MockExecMockRecorder struct {
	mock *MockExec
}

// NewMockExec creates a new mock instance.
func NewMockExec(ctrl *gomock.Controller) *MockExec {
	mock := &MockExec{ctrl: ctrl}
	mock.recorder = &MockExecMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExec) EXPECT() *MockExecMockRecorder {
	return m.recorder
}

// Command mocks base method.
func (m *MockExec) Command(name string, arg ...string) *exec.Cmd {
	m.ctrl.T.Helper()
	varargs := []interface{}{name}
	for _, a := range arg {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Command", varargs...)
	ret0, _ := ret[0].(*exec.Cmd)
	return ret0
}

// Command indicates an expected call of Command.
func (mr *MockExecMockRecorder) Command(name interface{}, arg ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name}, arg...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Command", reflect.TypeOf((*MockExec)(nil).Command), varargs...)
}
