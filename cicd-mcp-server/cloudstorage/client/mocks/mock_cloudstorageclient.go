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

package mocks

import (
	"context"
	"os"

	"cloud.google.com/go/iam"
)

// MockCloudStorageClient is a mock of CloudStorageClient interface.
type MockCloudStorageClient struct {
	GenerateUUIDFunc       func() string
	ListBucketsFunc        func(ctx context.Context, projectID string) ([]string, error)
	CheckBucketExistsFunc  func(ctx context.Context, bucketName string) error
	CreateBucketFunc       func(ctx context.Context, projectID, region, bucketName string) error
	UploadFileFunc         func(ctx context.Context, bucketName, objectName string, file *os.File) error
	CheckObjectExistsFunc  func(ctx context.Context, bucketName, objectName string) error
	GetBucketIamPolicyFunc func(ctx context.Context, bucketName string) (*iam.Policy, error)
	DeleteBucketFunc       func(ctx context.Context, bucketName string) error
	DeleteObjectsFunc      func(ctx context.Context, bucketName string) error
}

func (m *MockCloudStorageClient) GenerateUUID() string {
	return m.GenerateUUIDFunc()
}

// ListBuckets mocks the ListBuckets method.
func (m *MockCloudStorageClient) ListBuckets(ctx context.Context, projectID string) ([]string, error) {
	return m.ListBucketsFunc(ctx, projectID)
}

// CheckObjectExists mocks the CheckObjectExists method.
func (m *MockCloudStorageClient) CheckObjectExists(ctx context.Context, bucketName, objectName string) error {
	return m.CheckObjectExistsFunc(ctx, bucketName, objectName)
}

// GetBucketIamPolicy mocks the GetBucketIamPolicy method.
func (m *MockCloudStorageClient) GetBucketIamPolicy(ctx context.Context, bucketName string) (*iam.Policy, error) {
	return m.GetBucketIamPolicyFunc(ctx, bucketName)
}

// DeleteBucket mocks the DeleteBucket method.
func (m *MockCloudStorageClient) DeleteBucket(ctx context.Context, bucketName string) error {
	return m.DeleteBucketFunc(ctx, bucketName)
}

// DeleteObjects mocks the DeleteObjects method.
func (m *MockCloudStorageClient) DeleteObjects(ctx context.Context, bucketName string) error {
	return m.DeleteObjectsFunc(ctx, bucketName)
}

// CheckBucketExists mocks the CheckBucketExists method.
func (m *MockCloudStorageClient) CheckBucketExists(ctx context.Context, bucketName string) error {
	return m.CheckBucketExistsFunc(ctx, bucketName)
}

// CreateBucket mocks the CreateBucket method.
func (m *MockCloudStorageClient) CreateBucket(ctx context.Context, projectID, region, bucketName string) error {
	return m.CreateBucketFunc(ctx, projectID, region, bucketName)
}

// UploadFile mocks the UploadFile method.
func (m *MockCloudStorageClient) UploadFile(ctx context.Context, bucketName, objectName string, file *os.File) error {
	return m.UploadFileFunc(ctx, bucketName, objectName, file)
}
