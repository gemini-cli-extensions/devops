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

//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source=interface.go -destination=mocks/mock_interface.go -package=mocks
package cloudstorage

import (
	"context"
)

// GRPCClient is an interface for interacting with Google Cloud Storage.
type GRPClient interface {
	// CreateBucket creates a new GCS bucket.
	CreateBucket(ctx context.Context, projectID, bucketName string) error
	// UploadFile uploads a file to a GCS bucket.
	UploadFile(ctx context.Context, projectID, bucketName, objectName, filePath string) error
	// ReadFile reads a file from a GCS bucket.
	ReadFile(ctx context.Context, bucketName, objectName string) ([]byte, error)
	// UploadDirectory uploads a directory to a GCS bucket.
	UploadDirectory(ctx context.Context, projectID, bucketName, destinationDir, sourcePath string) error
}
