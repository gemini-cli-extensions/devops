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

package cloudrun

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"devops-mcp-server/cloudrun/client/mocks"

	cloudrunpb "cloud.google.com/go/run/apiv2/runpb"
)

func TestListServicesTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"

	tests := []struct {
		name                   string
		args                   ListServicesArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
		expectedServices       []*cloudrunpb.Service
	}{
		{
			name: "Success with no services",
			args: ListServicesArgs{
				ProjectID: projectID,
				Location:  location,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.ListServicesFunc = func(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error) {
					return []*cloudrunpb.Service{}, nil
				}
			},
			expectErr:        false,
			expectedServices: []*cloudrunpb.Service{},
		},
		{
			name: "Success with one service",
			args: ListServicesArgs{
				ProjectID: projectID,
				Location:  location,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.ListServicesFunc = func(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error) {
					return []*cloudrunpb.Service{{Name: "service-1"}}, nil
				}
			},
			expectErr:        false,
			expectedServices: []*cloudrunpb.Service{{Name: "service-1"}},
		},
		{
			name: "Success with multiple services",
			args: ListServicesArgs{
				ProjectID: projectID,
				Location:  location,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.ListServicesFunc = func(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error) {
					return []*cloudrunpb.Service{{Name: "service-1"}, {Name: "service-2"}}, nil
				}
			},
			expectErr:        false,
			expectedServices: []*cloudrunpb.Service{{Name: "service-1"}, {Name: "service-2"}},
		},
		{
			name: "Failure",
			args: ListServicesArgs{
				ProjectID: projectID,
				Location:  location,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.ListServicesFunc = func(ctx context.Context, projectID, location string) ([]*cloudrunpb.Service, error) {
					return nil, errors.New("error listing services")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to list services: error listing services",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.MockCloudRunClient{}
			tc.setupMocks(mockClient)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addListServicesTool(server, mockClient)

			_, result, err := listServicesToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("listServicesToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("listServicesToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}

			if !tc.expectErr {
				services, ok := result.([]*cloudrunpb.Service)
				if !ok {
					t.Errorf("listServicesToolFunc() result = %T, want %T", result, []*cloudrunpb.Service{})
				}
				if len(services) != len(tc.expectedServices) {
					t.Errorf("listServicesToolFunc() len(services) = %d, want %d", len(services), len(tc.expectedServices))
				}
			}
		})
	}
}

func TestCreateServiceTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	serviceName := "test-service"
	imageURL := "gcr.io/test-project/test-image"
	port := int32(8080)

	tests := []struct {
		name                   string
		args                   DeployToCloudRunFromImageArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name: "Success",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
			},
			expectErr: false,
		},
		{
			name: "Success with preexisting service",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return nil, status.Errorf(codes.AlreadyExists, "error service already exists")
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.GetRevisionFunc = func(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
					return &cloudrunpb.Revision{}, nil
				}
			},
			expectErr: false,
		},
		{
			name: "Fail creating service",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return nil, errors.New("error creating service")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to create service: error creating service",
		},
		{
			name: "Fail to get service",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return nil, status.Errorf(codes.AlreadyExists, "error service already exists")
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, errors.New("error getting service")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to get service: error getting service",
		},
		{
			name: "Fail to get update prexisting service",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return nil, status.Errorf(codes.AlreadyExists, "error service already exists")
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, errors.New("error updating service")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to update service with new revision: error updating service",
		},
		{
			name: "Fail to get revision after updating service",
			args: DeployToCloudRunFromImageArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				ImageURL:    imageURL,
				Port:        port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.CreateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL string, port int32) (*cloudrunpb.Service, error) {
					return nil, status.Errorf(codes.AlreadyExists, "error service already exists")
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.GetRevisionFunc = func(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
					return &cloudrunpb.Revision{}, errors.New("error getting revision")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to get revision: error getting revision",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.MockCloudRunClient{}
			tc.setupMocks(mockClient)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addDeployToCloudRunFromImageTool(server, mockClient)

			_, _, err := deployToCloudRunFromImageToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("createServiceToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("createServiceToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}
		})
	}
}

func TestCreateServiceFromSourceTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	serviceName := "test-service"
	source := "/path/to/source"

	tests := []struct {
		name                   string
		args                   DeployToCloudRunFromSourceArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name: "Success",
			args: DeployToCloudRunFromSourceArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				Source:      source,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.DeployFromSourceFunc = func(ctx context.Context, projectID, location, serviceName, source string, port int32) error {
					return nil
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
			},
			expectErr: false,
		},
		{
			name: "Failed to deploy from source",
			args: DeployToCloudRunFromSourceArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				Source:      source,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.DeployFromSourceFunc = func(ctx context.Context, projectID, location, serviceName, source string, port int32) error {
					return errors.New("error deploying")
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to create service: error deploying",
		},
		{
			name: "Failed to get deployed service",
			args: DeployToCloudRunFromSourceArgs{
				ProjectID:   projectID,
				Location:    location,
				ServiceName: serviceName,
				Source:      source,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.DeployFromSourceFunc = func(ctx context.Context, projectID, location, serviceName, source string, port int32) error {
					return nil
				}
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return nil, errors.New("error getting service")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to get service: error getting service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.MockCloudRunClient{}
			tc.setupMocks(mockClient)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addDeployToCloudRunFromSourceTool(server, mockClient)

			_, _, err := deployToCloudRunFromSourceToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("createServiceFromSourceToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("createServiceFromSourceToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}
		})
	}
}
