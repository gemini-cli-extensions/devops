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

package artifactregistryiface_test

import (
	"context"
	"testing"

	"devops-mcp-server/artifactregistryiface/mocks"

	"github.com/golang/mock/gomock"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
)

func TestMockGRPClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockGRPClient(ctrl)

	// Test CreateRepository
	createReq := &artifactregistrypb.CreateRepositoryRequest{
		Parent: "projects/my-project/locations/us-central1",
		Repository: &artifactregistrypb.Repository{
			Format: artifactregistrypb.Repository_DOCKER,
		},
		RepositoryId: "my-repo",
	}

	createRepo := &artifactregistrypb.Repository{
		Name: "projects/my-project/locations/us-central1/repositories/my-repo",
	}

	mockCreateOp := mocks.NewMockCreateRepositoryOperation(ctrl)
	mockClient.EXPECT().CreateRepository(gomock.Any(), createReq).Return(mockCreateOp, nil)
	mockCreateOp.EXPECT().Wait(gomock.Any()).Return(createRepo, nil)

	gotCreateRepoOp, err := mockClient.CreateRepository(context.Background(), createReq)
	if err != nil {
		t.Fatalf("CreateRepository() err = %v, want nil", err)
	}

	gotCreateRepo, err := gotCreateRepoOp.Wait(context.Background())
	if err != nil {
		t.Fatalf("CreateRepositoryOperation.Wait() err = %v, want nil", err)
	}

	if gotCreateRepo.Name != createRepo.Name {
		t.Errorf("CreateRepository() = %v, want %v", gotCreateRepo, createRepo)
	}

	// Test GetRepository
	getReq := &artifactregistrypb.GetRepositoryRequest{
		Name: "projects/my-project/locations/us-central1/repositories/my-repo",
	}

	getRepo := &artifactregistrypb.Repository{
		Name: "projects/my-project/locations/us-central1/repositories/my-repo",
		Format: artifactregistrypb.Repository_DOCKER,
	}

	mockClient.EXPECT().GetRepository(gomock.Any(), getReq).Return(getRepo, nil)

	gotGetRepo, err := mockClient.GetRepository(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetRepository() err = %v, want nil", err)
	}

	if gotGetRepo.Name != getRepo.Name {
		t.Errorf("GetRepository() = %v, want %v", gotGetRepo, getRepo)
	}
}
