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

package cloudbuild

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"devops-mcp-server/cloudbuildiface"

	"github.com/golang/mock/gomock"
	cloudbuildv1 "google.golang.org/api/cloudbuild/v1"

	"devops-mcp-server/cloudbuildiface/mocks"
	cloudstorage_mocks "devops-mcp-server/cloudstorage/mocks"
)

func TestBuildContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuildsService := mocks.NewMockBuildsServiceAPI(ctrl)
	mockCreateCall := mocks.NewMockBuildsCreateCallAPI(ctrl)
	mockGCSClient := cloudstorage_mocks.NewMockGRPClient(ctrl)
	mockOperationsService := mocks.NewMockOperationsServiceAPI(ctrl)
	mockGetCall := mocks.NewMockOperationsGetCallAPI(ctrl)
	c := &Client{
		buildsService: mockBuildsService,
		gcsClient:     mockGCSClient,
		regionalOperationsServiceFactory: func(ctx context.Context, location string) (cloudbuildiface.OperationsServiceAPI, error) {
			return mockOperationsService, nil
		},
	}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	repository := "test-repo"
	imageName := "test-image"
	tag := "latest"
	tmpDir := t.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	startOperation := &cloudbuildv1.Operation{Name: "operations/test-op"}
	doneOperation := &cloudbuildv1.Operation{Name: "operations/test-op", Done: true}

	// Create a dummy Dockerfile
	if err := os.WriteFile(dockerfilePath, []byte("FROM busybox"), 0644); err != nil {
		t.Fatalf("Failed to create dummy Dockerfile: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		mockGCSClient.EXPECT().UploadFile(gomock.Any(), projectID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		mockBuildsService.EXPECT().Create(parent, gomock.Any()).Return(mockCreateCall)
		mockCreateCall.EXPECT().Context(ctx).Return(mockCreateCall)
		mockCreateCall.EXPECT().Do().Return(startOperation, nil)
		mockOperationsService.EXPECT().Get(startOperation.Name).Return(mockGetCall).Times(2)
		mockGetCall.EXPECT().Context(ctx).Return(mockGetCall).Times(2)
		mockGetCall.EXPECT().Do(gomock.Any()).Return(startOperation, nil)
		mockGetCall.EXPECT().Do(gomock.Any()).Return(doneOperation, nil)

		op, err := c.BuildContainer(ctx, projectID, location, repository, imageName, tag, dockerfilePath)
		if err != nil {
			t.Fatalf("BuildContainer() error = %v, want nil", err)
		}
		if !op.Done {
			t.Error("BuildContainer() did not return a done operation")
		}
	})
}

func TestCreateTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTriggersService := mocks.NewMockTriggersServiceAPI(ctrl)
	mockCreateCall := mocks.NewMockTriggersCreateCallAPI(ctrl)
	c := &Client{triggersService: mockTriggersService}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	triggerID := "test-trigger"
	repoLink := "test-repo-link"
	serviceAccount := "test-sa"
	branch := "main"

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	expectedTrigger := &cloudbuildv1.BuildTrigger{Name: triggerID}

	t.Run("success", func(t *testing.T) {
		mockTriggersService.EXPECT().Create(parent, gomock.Any()).Return(mockCreateCall)
		mockCreateCall.EXPECT().Context(ctx).Return(mockCreateCall)
		mockCreateCall.EXPECT().Do().Return(expectedTrigger, nil)

		trigger, err := c.CreateTrigger(ctx, projectID, location, triggerID, repoLink, serviceAccount, branch, "")
		if err != nil {
			t.Fatalf("CreateTrigger() error = %v, want nil", err)
		}
		if trigger.Name != triggerID {
			t.Errorf("CreateTrigger() got trigger name = %v, want %v", trigger.Name, triggerID)
		}
	})

	t.Run("error", func(t *testing.T) {
		mockTriggersService.EXPECT().Create(parent, gomock.Any()).Return(mockCreateCall)
		mockCreateCall.EXPECT().Context(ctx).Return(mockCreateCall)
		mockCreateCall.EXPECT().Do().Return(nil, errors.New("API error"))

		_, err := c.CreateTrigger(ctx, projectID, location, triggerID, repoLink, serviceAccount, branch, "")
		if err == nil {
			t.Error("expected an error, but got nil")
		}
	})

	t.Run("invalid args", func(t *testing.T) {
		_, err := c.CreateTrigger(ctx, projectID, location, triggerID, repoLink, serviceAccount, "", "")
		if err == nil {
			t.Error("expected an error when both branch and tag are empty, but got nil")
		}
	})
}

func TestRunTrigger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTriggersService := mocks.NewMockTriggersServiceAPI(ctrl)
	mockRunCall := mocks.NewMockTriggersRunCallAPI(ctrl)
	c := &Client{triggersService: mockTriggersService}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	triggerID := "test-trigger"
	name := fmt.Sprintf("projects/%s/locations/%s/triggers/%s", projectID, location, triggerID)
	expectedOp := &cloudbuildv1.Operation{Name: "operations/test-op"}

	t.Run("success", func(t *testing.T) {
		mockTriggersService.EXPECT().Run(name, gomock.Any()).Return(mockRunCall)
		mockRunCall.EXPECT().Context(ctx).Return(mockRunCall)
		mockRunCall.EXPECT().Do().Return(expectedOp, nil)

		op, err := c.RunTrigger(ctx, projectID, location, triggerID)
		if err != nil {
			t.Fatalf("RunTrigger() error = %v, want nil", err)
		}
		if op.Name != expectedOp.Name {
			t.Errorf("RunTrigger() got operation name = %v, want %v", op.Name, expectedOp.Name)
		}
	})

	t.Run("error", func(t *testing.T) {
		mockTriggersService.EXPECT().Run(name, gomock.Any()).Return(mockRunCall)
		mockRunCall.EXPECT().Context(ctx).Return(mockRunCall)
		mockRunCall.EXPECT().Do().Return(nil, errors.New("API error"))

		_, err := c.RunTrigger(ctx, projectID, location, triggerID)
		if err == nil {
			t.Error("expected an error, but got nil")
		}
	})
}

func TestListTriggers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTriggersService := mocks.NewMockTriggersServiceAPI(ctrl)
	mockListCall := mocks.NewMockTriggersListCallAPI(ctrl)
	c := &Client{triggersService: mockTriggersService}

	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
	expectedTriggers := []*cloudbuildv1.BuildTrigger{
		{Name: "trigger1"},
		{Name: "trigger2"},
	}
	expectedResp := &cloudbuildv1.ListBuildTriggersResponse{Triggers: expectedTriggers}

	t.Run("success", func(t *testing.T) {
		mockTriggersService.EXPECT().List(parent).Return(mockListCall)
		mockListCall.EXPECT().Context(ctx).Return(mockListCall)
		mockListCall.EXPECT().Do().Return(expectedResp, nil)

		triggers, err := c.ListTriggers(ctx, projectID, location)
		if err != nil {
			t.Fatalf("ListTriggers() error = %v, want nil", err)
		}
		if len(triggers) != 2 {
			t.Errorf("ListTriggers() got %d triggers, want 2", len(triggers))
		}
	})

	t.Run("error", func(t *testing.T) {
		mockTriggersService.EXPECT().List(parent).Return(mockListCall)
		mockListCall.EXPECT().Context(ctx).Return(mockListCall)
		mockListCall.EXPECT().Do().Return(nil, errors.New("API error"))

		_, err := c.ListTriggers(ctx, projectID, location)
		if err == nil {
			t.Error("expected an error, but got nil")
		}
	})
}
