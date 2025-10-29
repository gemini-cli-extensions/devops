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
)

// MockCloudStorageClient is a mock of CloudStorageClient interface.
type MockCloudStorageClient struct {
	CheckBucketExistsFunc func(ctx context.Context, bucketName string) error
	CreateBucketFunc func(ctx context.Context, projectID, bucketName string) error
	UploadFileFunc func(ctx context.Context, bucketName, objectName string, file *os.File) error
}

// CheckBucketExists mocks the CheckBucketExists method.
func (m *MockCloudStorageClient) CheckBucketExists(ctx context.Context, bucketName string) error {
	return m.CheckBucketExistsFunc(ctx, bucketName)
}

// CreateBucket mocks the CreateBucket method.
func (m *MockCloudStorageClient) CreateBucket(ctx context.Context, projectID, bucketName string) error {
	return m.CreateBucketFunc(ctx, projectID, bucketName)
}

// UploadFile mocks the UploadFile method.
func (m *MockCloudStorageClient) UploadFile(ctx context.Context, bucketName, objectName string, file *os.File) error {
	return m.UploadFileFunc(ctx, bucketName, objectName, file)
}
