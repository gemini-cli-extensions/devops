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

func TestCreateServiceTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	serviceName := "test-service"
	imageURL := "gcr.io/test-project/test-image"
	port := int32(8080)

	tests := []struct {
		name                   string
		args                   CreateServiceArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name: "Success",
			args: CreateServiceArgs{
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
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
			},
			expectErr: false,
		},
		{
			name: "Success with preexisting service",
			args: CreateServiceArgs{
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
			},
			expectErr: false,
		},
		{
			name: "Fail creating service",
			args: CreateServiceArgs{
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
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to create service: error creating service",
		},
		{
			name: "Fail to get service",
			args: CreateServiceArgs{
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
			args: CreateServiceArgs{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.MockCloudRunClient{}
			tc.setupMocks(mockClient)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addCreateServiceTool(server, mockClient)

			_, _, err := createServiceToolFunc(ctx, nil, tc.args)

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

func TestCreateRevisionTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	serviceName := "test-service"
	revisionName := "test-revision"
	imageURL := "gcr.io/test-project/test-image"
	port := int32(8080)

	tests := []struct {
		name                   string
		args                   CreateRevisionArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name: "Success",
			args: CreateRevisionArgs{
				ProjectID:    projectID,
				Location:     location,
				ServiceName:  serviceName,
				ImageURL:     imageURL,
				RevisionName: revisionName,
				Port:         port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
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
			name: "Fail to get service",
			args: CreateRevisionArgs{
				ProjectID:    projectID,
				Location:     location,
				ServiceName:  serviceName,
				ImageURL:     imageURL,
				RevisionName: revisionName,
				Port:         port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, errors.New("error getting service")
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.GetRevisionFunc = func(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
					return &cloudrunpb.Revision{}, nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to get service: error getting service",
		},
		{
			name: "Fail to update service",
			args: CreateRevisionArgs{
				ProjectID:    projectID,
				Location:     location,
				ServiceName:  serviceName,
				ImageURL:     imageURL,
				RevisionName: revisionName,
				Port:         port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
				mockClient.GetServiceFunc = func(ctx context.Context, projectID, location, serviceName string) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, nil
				}
				mockClient.UpdateServiceFunc = func(ctx context.Context, projectID, location, serviceName, imageURL, revisionName string, port int32, service *cloudrunpb.Service) (*cloudrunpb.Service, error) {
					return &cloudrunpb.Service{}, errors.New("error updating service")
				}
				mockClient.GetRevisionFunc = func(ctx context.Context, service *cloudrunpb.Service) (*cloudrunpb.Revision, error) {
					return &cloudrunpb.Revision{}, nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to update service with new revision: error updating service",
		},
		{
			name: "Fail to get revision",
			args: CreateRevisionArgs{
				ProjectID:    projectID,
				Location:     location,
				ServiceName:  serviceName,
				ImageURL:     imageURL,
				RevisionName: revisionName,
				Port:         port,
			},
			setupMocks: func(mockClient *mocks.MockCloudRunClient) {
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
			addCreateRevisionTool(server, mockClient)

			_, _, err := createRevisionToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("createRevisionToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("createRevisionToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
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
		args                   CreateServiceFromSourceArgs
		setupMocks             func(*mocks.MockCloudRunClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name: "Success",
			args: CreateServiceFromSourceArgs{
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
			args: CreateServiceFromSourceArgs{
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
			args: CreateServiceFromSourceArgs{
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
			addCreateServiceFromSourceTool(server, mockClient)

			_, _, err := createServiceFromSourceToolFunc(ctx, nil, tc.args)

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
