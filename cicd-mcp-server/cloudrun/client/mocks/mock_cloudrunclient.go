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

	cloudrunpb "cloud.google.com/go/run/apiv2/runpb"
)

// MockCloudRunClient is a mock of CloudRunClient interface.
type MockCloudRunClient struct {
	GetServiceFunc       func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error)
	ListServicesFunc     func(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error)
	CreateServiceFunc    func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error)
	UpdateServiceFunc    func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error)
	GetRevisionFunc      func(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error)
	DeployFromSourceFunc func(ctx context.Context, projectID, location, serviceName, source string, port int32, allowPublicAccess bool) error
	DeleteServiceFunc    func(ctx context.Context, projectID, location, serviceName string) error
	SetServiceAccessFunc func(ctx context.Context, serviceName string, allowPublicAccess bool) error
}

// DeleteService mocks the DeleteService method.
func (m *MockCloudRunClient) DeleteService(ctx context.Context, projectID, location, serviceName string) error {
	return m.DeleteServiceFunc(ctx, projectID, location, serviceName)
}

// GetService mocks the GetService method.
func (m *MockCloudRunClient) GetService(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
	return m.GetServiceFunc(ctx, projectID, location, serviceName)
}

// ListServices mocks the ListServices method.
func (m *MockCloudRunClient) ListServices(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error) {
	return m.ListServicesFunc(ctx, projectID, location)
}

// CreateService mocks the CreateService method.
func (m *MockCloudRunClient) CreateService(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
	return m.CreateServiceFunc(ctx, projectID, location, serviceName, imageURL, port)
}

// UpdateService mocks the UpdateService method.
func (m *MockCloudRunClient) UpdateService(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
	return m.UpdateServiceFunc(ctx, projectID, location, serviceName, imageURL, revisionName, port, service)
}

// GetRevision mocks the GetRevision method.
func (m *MockCloudRunClient) GetRevision(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
	return m.GetRevisionFunc(ctx, service)
}

// DeployFromSource mocks the DeployFromSource method.
func (m *MockCloudRunClient) DeployFromSource(ctx context.Context, projectID, location, serviceName, source string, port int32, allowPublicAccess bool) error {
	return m.DeployFromSourceFunc(ctx, projectID, location, serviceName, source, port, allowPublicAccess)
}

func (m *MockCloudRunClient) SetServiceAccess(ctx context.Context, serviceName string, allowPublicAccess bool) error {
	return m.SetServiceAccessFunc(ctx, serviceName, allowPublicAccess)
}
