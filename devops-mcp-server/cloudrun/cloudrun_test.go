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
	"os"
	"os/exec"
	"testing"

	"github.com/golang/mock/gomock"

	"devops-mcp-server/cloudrun/mocks"
)

const (
	projectID   = "test-project"
	location    = "us-central1"
	serviceName = "test-service"
	imageURL    = "gcr.io/test-project/test-image"
	source      = "/path/to/source"
)

func TestDeployFromImage(t *testing.T) {
	// Successful deployment
	t.Run("Success", func(t *testing.T) {
		cmd := helperSuccess(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--image", imageURL, "--project", projectID, "--region", location, "--format", "json", "--quiet").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromImage(context.Background(), projectID, location, serviceName, imageURL, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Successful deployment with port
	t.Run("SuccessWithPort", func(t *testing.T) {
		cmd := helperSuccess(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--image", imageURL, "--project", projectID, "--region", location, "--format", "json", "--quiet", "--port", "8080").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromImage(context.Background(), projectID, location, serviceName, imageURL, 8080)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Failed deployment
	t.Run("Failure", func(t *testing.T) {
		cmd := helperFailure(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--image", imageURL, "--project", projectID, "--region", location, "--format", "json", "--quiet").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromImage(context.Background(), projectID, location, serviceName, imageURL, 0)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestDeployFromSource(t *testing.T) {
	// Successful deployment
	t.Run("Success", func(t *testing.T) {
		cmd := helperSuccess(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--project", projectID, "--region", location, "--source", source, "--format", "json", "--quiet").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromSource(context.Background(), projectID, location, serviceName, source, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Successful deployment with port
	t.Run("SuccessWithPort", func(t *testing.T) {
		cmd := helperSuccess(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--project", projectID, "--region", location, "--source", source, "--format", "json", "--quiet", "--port", "8080").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromSource(context.Background(), projectID, location, serviceName, source, 8080)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// Failed deployment
	t.Run("Failure", func(t *testing.T) {
		cmd := helperFailure(t)
		ctrl := gomock.NewController(t)
		mockExecer := mocks.NewMockExec(ctrl)
		mockExecer.EXPECT().Command("gcloud", "run", "deploy", serviceName, "--project", projectID, "--region", location, "--source", source, "--format", "json", "--quiet").Return(cmd)

		c := &Client{execer: mockExecer}
		err := c.DeployFromSource(context.Background(), projectID, location, serviceName, source, 0)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func helperSuccess(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessSuccess")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS_SUCCESS=1"}
	return cmd
}

func helperFailure(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessFailure")
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS_FAILURE=1"}
	return cmd
}

func TestHelperProcessSuccess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS_SUCCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestHelperProcessFailure(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS_FAILURE") != "1" {
		return
	}
	os.Exit(1)
}