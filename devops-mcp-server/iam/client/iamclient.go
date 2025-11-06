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

package iamclient

import (
	"context"
	"fmt"

	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
	iamv1 "google.golang.org/api/iam/v1"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	iamClientKey contextKey = "iamClient"
)

// ClientFrom returns the IAMClient stored in the context, if any.
func ClientFrom(ctx context.Context) (IAMClient, bool) {
	client, ok := ctx.Value(iamClientKey).(IAMClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided IAMClient.
func ContextWithClient(ctx context.Context, client IAMClient) context.Context {
	return context.WithValue(ctx, iamClientKey, client)
}

// ServiceAccountList defines a struct to wrap a list of service accounts.
type ServiceAccountList struct {
	Items []*iamv1.ServiceAccount `json:"items"`
}

// RoleBindingList defines a struct to wrap a list of role bindings.
type RoleBindingList struct {
	Items []string `json:"items"`
}

// Client is an interface for interacting with the IAM API.
type IAMClient interface {
	CreateServiceAccount(ctx context.Context, projectID, displayName, accountID string) (*iamv1.ServiceAccount, error)
	AddIAMRoleBinding(ctx context.Context, resourceID, role, member string) (*cloudresourcemanagerv1.Policy, error)
	ListServiceAccounts(ctx context.Context, projectID string) (*ServiceAccountList, error)
	GetIAMRoleBinding(ctx context.Context, projectID, serviceAccountEmail string) (*RoleBindingList, error)
}

// clientImpl is a client for interacting with the IAM API.
type IAMClientImpl struct {
	iamService *iamv1.Service
	crmService *cloudresourcemanagerv1.Service
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (IAMClient, error) {
	iamService, err := iamv1.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create iam service: %v", err)
	}
	crmService, err := cloudresourcemanagerv1.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud resource manager service: %v", err)
	}

	return &IAMClientImpl{iamService: iamService, crmService: crmService}, nil
}

// CreateServiceAccount creates a new Google Cloud Platform service account.
func (c *IAMClientImpl) CreateServiceAccount(ctx context.Context, projectID, displayName, accountID string) (*iamv1.ServiceAccount, error) {
	projectPath := fmt.Sprintf("projects/%s", projectID)
	req := &iamv1.CreateServiceAccountRequest{
		AccountId: accountID,
		ServiceAccount: &iamv1.ServiceAccount{
			DisplayName: displayName,
		},
	}

	return c.iamService.Projects.ServiceAccounts.Create(projectPath, req).Context(ctx).Do()
}

// AddIAMRoleBinding adds an IAM role binding to a Google Cloud Platform resource.
func (c *IAMClientImpl) AddIAMRoleBinding(ctx context.Context, resourceID, role, member string) (*cloudresourcemanagerv1.Policy, error) {
	policy, err := c.crmService.Projects.GetIamPolicy(resourceID, &cloudresourcemanagerv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get iam policy: %v", err)
	}

	policy.Bindings = append(policy.Bindings, &cloudresourcemanagerv1.Binding{
		Role:    role,
		Members: []string{member},
	})

	setPolicyRequest := &cloudresourcemanagerv1.SetIamPolicyRequest{
		Policy: policy,
	}

	return c.crmService.Projects.SetIamPolicy(resourceID, setPolicyRequest).Context(ctx).Do()
}

// ListServiceAccounts lists all service accounts in a project.
func (c *IAMClientImpl) ListServiceAccounts(ctx context.Context, projectID string) (*ServiceAccountList, error) {
	parent := fmt.Sprintf("projects/%s", projectID)

	resp, err := c.iamService.Projects.ServiceAccounts.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %v", err)
	}

	return &ServiceAccountList{Items: resp.Accounts}, nil
}

// GetIAMRoleBinding gets the IAM role bindings for a service account.
func (c *IAMClientImpl) GetIAMRoleBinding(ctx context.Context, projectID, serviceAccountEmail string) (*RoleBindingList, error) {
	policy, err := c.crmService.Projects.GetIamPolicy(projectID, &cloudresourcemanagerv1.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get iam policy: %v", err)
	}

	var roles []string

	for _, binding := range policy.Bindings {
		for _, member := range binding.Members {
			if member == fmt.Sprintf("serviceAccount:%s", serviceAccountEmail) {
				roles = append(roles, binding.Role)
			}
		}
	}

	return &RoleBindingList{Items: roles}, nil
}
