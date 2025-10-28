// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mocks

import (
	"context"
	"fmt"

	iamclient "devops-mcp-server/iam/client"

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	iamv1 "google.golang.org/api/iam/v1"
)

// MockIAMClient is a mock of IAMClient interface.
type MockIAMClient struct {
	AddIAMRoleBindingFunc func(ctx context.Context, resourceID, role, member string) (*cloudresourcemanagerv1.Policy, error)
}

// CreateServiceAccount mocks the CreateServiceAccount method.
func (m *MockIAMClient) CreateServiceAccount(ctx context.Context, projectID, displayName, accountID string) (*iamv1.ServiceAccount, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddIAMRoleBinding mocks the AddIAMRoleBinding method.
func (m *MockIAMClient) AddIAMRoleBinding(ctx context.Context, resourceID, role, member string) (*cloudresourcemanagerv1.Policy, error) {
	return m.AddIAMRoleBindingFunc(ctx, resourceID, role, member)
}

// ListServiceAccounts mocks the ListServiceAccounts method.
func (m *MockIAMClient) ListServiceAccounts(ctx context.Context, projectID string) (*iamclient.ListResult[*iamv1.ServiceAccount], error) {
	return nil, fmt.Errorf("not implemented")
}

// GetIAMRoleBinding mocks the GetIAMRoleBinding method.
func (m *MockIAMClient) GetIAMRoleBinding(ctx context.Context, projectID, serviceAccountEmail string) (*iamclient.ListResult[string], error) {
	return nil, fmt.Errorf("not implemented")
}
