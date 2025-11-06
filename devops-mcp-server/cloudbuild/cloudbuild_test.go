package cloudbuild

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	iamclientMock "devops-mcp-server/iam/client/mocks"
	resourcemanagerclientMock "devops-mcp-server/resourcemanager/client/mocks"
)

func TestSetPermissionsForCloudBuildSA(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRMClient := resourcemanagerclientMock.NewMockResourcemanagerClient(ctrl)
	mockIAMClient := iamclientMock.NewMockIAMClient(ctrl)

	ctx := context.Background()
	projectID := "test-project"
	serviceAccount := "test-sa@example.com"

	t.Run("with service account", func(t *testing.T) {
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/developerconnect.tokenAccessor", serviceAccount).Return(nil, nil)

		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.NoError(t, err)
		assert.Equal(t, serviceAccount, resolvedSA)
	})

	t.Run("without service account", func(t *testing.T) {
		projectNumber := int64(12345)
		expectedSA := fmt.Sprintf("%d-compute@developer.gserviceaccount.com", projectNumber)

		mockRMClient.EXPECT().ToProjectNumber(ctx, projectID).Return(projectNumber, nil)
		mockIAMClient.EXPECT().AddIAMRoleBinding(ctx, fmt.Sprintf("projects/%s", projectID), "roles/developerconnect.tokenAccessor", expectedSA).Return(nil, nil)

		resolvedSA, err := setPermissionsForCloudBuildSA(ctx, projectID, "", mockRMClient, mockIAMClient)
		assert.NoError(t, err)
		assert.Equal(t, expectedSA, resolvedSA)
	})

	t.Run("iam error", func(t *testing.T) {
		mockIAMClient.EXPECT().AddIAMRoleBinding(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error"))

		_, err := setPermissionsForCloudBuildSA(ctx, projectID, serviceAccount, mockRMClient, mockIAMClient)
		assert.Error(t, err)
	})

	t.Run("resourcemanager error", func(t *testing.T) {
		mockRMClient.EXPECT().ToProjectNumber(gomock.Any(), gomock.Any()).Return(int64(0), fmt.Errorf("some error"))

		_, err := setPermissionsForCloudBuildSA(ctx, projectID, "", mockRMClient, mockIAMClient)
		assert.Error(t, err)
	})
}
