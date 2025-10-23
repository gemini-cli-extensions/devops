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

package artifactregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	artifactregistrypb "google.golang.org/genproto/googleapis/devtools/artifactregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"devops-mcp-server/artifactregistryiface/mocks"
)

func TestGetRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGRPClient := mocks.NewMockGRPClient(ctrl)
	client, err := NewClient(mockGRPClient)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	repositoryID := "test-repo"

	getRequest := &artifactregistrypb.GetRepositoryRequest{
		Name: "projects/test-project/locations/us-central1/repositories/test-repo",
	}
	repo := &artifactregistrypb.Repository{
		Name:   "projects/test-project/locations/us-central1/repositories/test-repo",
		Format: artifactregistrypb.Repository_DOCKER,
	}

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "Repository Found",
			setupMock: func() {
				mockGRPClient.EXPECT().GetRepository(ctx, getRequest).Return(repo, nil)
			},
			expectErr: false,
		},
		{
			name: "Repository Not Found",
			setupMock: func() {
				mockGRPClient.EXPECT().GetRepository(ctx, getRequest).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			_, err := client.GetRepository(ctx, projectID, location, repositoryID)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetRepository() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestCreateRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGRPClient := mocks.NewMockGRPClient(ctrl)
	client, err := NewClient(mockGRPClient)
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	repositoryID := "test-repo"
	format := "DOCKER"

	getRequest := &artifactregistrypb.GetRepositoryRequest{
		Name: "projects/test-project/locations/us-central1/repositories/test-repo",
	}

	createRequest := &artifactregistrypb.CreateRepositoryRequest{
		Parent:       "projects/test-project/locations/us-central1",
		RepositoryId: "test-repo",
		Repository: &artifactregistrypb.Repository{
			Format: artifactregistrypb.Repository_DOCKER,
		},
	}
	repo := &artifactregistrypb.Repository{
		Name:   "projects/test-project/locations/us-central1/repositories/test-repo",
		Format: artifactregistrypb.Repository_DOCKER,
	}

	tests := []struct {
		name          string
		setupMock     func()
		expectErr     bool
		expectedError string
	}{
		{
			name: "Repository Exists",
			setupMock: func() {
				mockGRPClient.EXPECT().GetRepository(ctx, getRequest).Return(repo, nil)
			},
			expectErr: false,
		},
		{
			name: "Repository Does Not Exist, Create New",
			setupMock: func() {
				mockGRPClient.EXPECT().GetRepository(ctx, getRequest).Return(nil, status.Error(codes.NotFound, "not found"))
				mockCreateOp := mocks.NewMockCreateRepositoryOperation(ctrl)
				mockGRPClient.EXPECT().CreateRepository(ctx, createRequest).Return(mockCreateOp, nil)
				mockCreateOp.EXPECT().Wait(ctx).Return(repo, nil)
			},
			expectErr: false,
		},
		{
			name: "GetRepository Returns Unexpected Error",
			setupMock: func() {
				mockGRPClient.EXPECT().GetRepository(ctx, getRequest).Return(nil, errors.New("unexpected error"))
			},
			expectErr:     true,
			expectedError: "failed to get repository: unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			_, err := client.CreateRepository(ctx, projectID, location, repositoryID, format)
			if (err != nil) != tt.expectErr {
				t.Errorf("CreateRepository() error = %v, expectErr %v", err, tt.expectErr)
			}
			if tt.expectErr && err.Error() != tt.expectedError {
				t.Errorf("CreateRepository() error = %v, expectedError %v", err.Error(), tt.expectedError)
			}
		})
	}
}