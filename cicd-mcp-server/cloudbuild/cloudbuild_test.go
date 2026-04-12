// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudbuild

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	iamclientMock "cicd-mcp-server/iam/client/mocks"
	resourcemanagerclientMock "cicd-mcp-server/resourcemanager/client/mocks"
)

func TestSetPermissionsForCloudBuildSA(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRMClient := resourcemanagerclientMock.NewMockResourcemanagerClient(ctrl)
	mockIAMClient := iamclientMock.NewMockIAMClient(ctrl)

	ctx := context.Background()
	projectID := "test-project"
	serviceAccount := "serviceAccount:test-sa@example.com"
	serviceAccountWOPrefix := "test-sa@example.com"

	projectNumber := int64(12345)
	gcbSARoles := []string{
		"roles/artifactregistry.writer",
		"roles/cloudbuild.builds.editor",
		"roles/developerconnect.tokenAccessor",
		"roles/logging.logWriter",
		"roles/run.developer",
		"roles/serviceusage.serviceUsageConsumer",
		"roles/storage.admin",
	}
	dcP4sa := fmt.Sprintf("serviceAccount:service-%d@gcp-sa-devconnect.iam.gserviceaccount.com", projectNumber)
	gcbP4sa := fmt.Sprintf("serviceAccount:service-%d@gcp-sa-cloudbuild.iam.gserviceaccount.com", projectNumber)
	gcbP4saRoles := []string{
		"roles/cloudbuild.serviceAgent",
		"roles/developerconnect.tokenAccessor",
	}
	defaultComputeSA := fmt.Sprintf("%d-compute@developer.gserviceaccount.com", projectNumber)
	saResource := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, defaultComputeSA)
	userSAResource := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, "test-sa@example.com")

	t.Run("with service account", func(t *testing.T) {
		for _, r := range gcbSARoles {
			mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, serviceAccount).Return(nil, nil)
		}
		mockRMClient.EXPECT().ToProjectNumber(ctx, projectID).Return(projectNumber, nil)
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/secretmanager.admin", dcP4sa).Return(nil, nil)
		for _, r := range gcbP4saRoles {
			mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, gcbP4sa).Return(nil, nil)
		}
		mockIAMClient.EXPECT().AddServiceAccountRoleBinding(ctx, saResource, "roles/iam.serviceAccountUser", gcbP4sa).Return(nil, nil)
		mockIAMClient.EXPECT().AddServiceAccountRoleBinding(ctx, userSAResource, "roles/iam.serviceAccountUser", gcbP4sa).Return(nil, nil)

		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.NoError(t, err)
		assert.Equal(t, serviceAccount, resolvedSA)
	})

	t.Run("with service account, no prefix", func(t *testing.T) {
		for _, r := range gcbSARoles {
			mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, serviceAccount).Return(nil, nil)
		}
		mockRMClient.EXPECT().ToProjectNumber(ctx, projectID).Return(projectNumber, nil)
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/secretmanager.admin", dcP4sa).Return(nil, nil)
		for _, r := range gcbP4saRoles {
			mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, gcbP4sa).Return(nil, nil)
		}
		mockIAMClient.EXPECT().AddServiceAccountRoleBinding(ctx, saResource, "roles/iam.serviceAccountUser", gcbP4sa).Return(nil, nil)
		mockIAMClient.EXPECT().AddServiceAccountRoleBinding(ctx, userSAResource, "roles/iam.serviceAccountUser", gcbP4sa).Return(nil, nil)

		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccountWOPrefix, mockRMClient, mockIAMClient)
		assert.NoError(t, err)
		assert.Equal(t, serviceAccount, resolvedSA)
	})

	t.Run("no service account provided fails", func(t *testing.T) {
		// Since fallback is removed, providing an empty SA should result in an invalid SA prefix 
		// and fail during IAM role binding.
		mockIAMClient.EXPECT().AddIAMRoleBinding(gomock.Any(), gomock.Any(), gomock.Any(), "serviceAccount:").Return(nil, fmt.Errorf("invalid member"))

		_, err := setPermissionsForCloudBuildSA(ctx, projectID, "", mockRMClient, mockIAMClient)
		assert.Error(t, err)
	})

	t.Run("iam error", func(t *testing.T) {
		mockIAMClient.EXPECT().AddIAMRoleBinding(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error"))

		_, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.Error(t, err)
	})

	t.Run("resourcemanager error", func(t *testing.T) {
		// Mock success for the first set of roles
		for _, r := range gcbSARoles {
			mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), r, serviceAccount).Return(nil, nil)
		}
		// Mock failure for ToProjectNumber
		mockRMClient.EXPECT().ToProjectNumber(ctx, projectID).Return(int64(0), fmt.Errorf("some error"))

		_, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.Error(t, err)
	})
}

func TestIsValidServiceAccount(t *testing.T) {
	tests := []struct {
		name string
		sa   string
		want bool
	}{
		{"valid sa", "serviceAccount:test-sa@project.iam.gserviceaccount.com", true},
		{"valid sa with dashes", "serviceAccount:my-sa-123@my-project-456.iam.gserviceaccount.com", true},
		{"missing prefix", "test-sa@project.iam.gserviceaccount.com", false},
		{"wrong prefix", "user:test-sa@project.iam.gserviceaccount.com", false},
		{"missing domain", "serviceAccount:test-sa@project", false},
		{"wrong domain", "serviceAccount:test-sa@project.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidServiceAccount(tt.sa); got != tt.want {
				t.Errorf("IsValidServiceAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
