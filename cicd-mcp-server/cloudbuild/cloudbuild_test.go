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

	t.Run("with service account", func(t *testing.T) {
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/developerconnect.tokenAccessor", serviceAccount).Return(nil, nil)

		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.NoError(t, err)
		assert.Equal(t, serviceAccount, resolvedSA)
	})

	t.Run("with service account, no prefix", func(t *testing.T) {
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/developerconnect.tokenAccessor", serviceAccount).Return(nil, nil)

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
}
