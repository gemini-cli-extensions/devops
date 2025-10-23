package cloudstorage

import (
	"context"
	"devops-mcp-server/cloudstorage/mocks"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestGCSClient_CreateBucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSClient := mocks.NewMockGRPClient(ctrl)
	ctx := context.Background()
	projectID := "test-project"
	bucketName := "test-bucket"

	t.Run("success", func(t *testing.T) {
		mockGCSClient.EXPECT().CreateBucket(ctx, projectID, bucketName).Return(nil)
		err := mockGCSClient.CreateBucket(ctx, projectID, bucketName)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("failed to create bucket")
		mockGCSClient.EXPECT().CreateBucket(ctx, projectID, bucketName).Return(expectedErr)
		err := mockGCSClient.CreateBucket(ctx, projectID, bucketName)
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestGCSClient_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSClient := mocks.NewMockGRPClient(ctrl)
	ctx := context.Background()
	projectID := "test-project"
	bucketName := "test-bucket"
	objectName := "test-object"
	filePath := "/path/to/test/file.txt"

	t.Run("success", func(t *testing.T) {
		mockGCSClient.EXPECT().UploadFile(ctx, projectID, bucketName, objectName, filePath).Return(nil)
		err := mockGCSClient.UploadFile(ctx, projectID, bucketName, objectName, filePath)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("failed to upload file")
		mockGCSClient.EXPECT().UploadFile(ctx, projectID, bucketName, objectName, filePath).Return(expectedErr)
		err := mockGCSClient.UploadFile(ctx, projectID, bucketName, objectName, filePath)
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestGCSClient_ReadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSClient := mocks.NewMockGRPClient(ctrl)
	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-object"

	t.Run("success", func(t *testing.T) {
		expectedData := []byte("test data")
		mockGCSClient.EXPECT().ReadFile(ctx, bucketName, objectName).Return(expectedData, nil)
		data, err := mockGCSClient.ReadFile(ctx, bucketName, objectName)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(data, expectedData) {
			t.Errorf("expected data %v, got %v", expectedData, data)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("failed to read file")
		mockGCSClient.EXPECT().ReadFile(ctx, bucketName, objectName).Return(nil, expectedErr)
		_, err := mockGCSClient.ReadFile(ctx, bucketName, objectName)
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestGCSClient_UploadDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGCSClient := mocks.NewMockGRPClient(ctrl)
	ctx := context.Background()
	projectID := "test-project"
	bucketName := "test-bucket"
	destinationDir := "test-destination"
	sourcePath := "/path/to/test/dir"

	t.Run("success", func(t *testing.T) {
		mockGCSClient.EXPECT().UploadDirectory(ctx, projectID, bucketName, destinationDir, sourcePath).Return(nil)
		err := mockGCSClient.UploadDirectory(ctx, projectID, bucketName, destinationDir, sourcePath)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
