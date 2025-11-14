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

package cloudstorage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	csmocks "devops-mcp-server/cloudstorage/client/mocks"

	storage "cloud.google.com/go/storage"
)

func TestAddListBucketsTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"

	tests := []struct {
		name                   string
		args                   ListBucketsArgs
		setupMocks             func(*csmocks.MockCloudStorageClient)
		expectErr              bool
		expectedErrorSubstring string
		expectedResult         []string
	}{
		{
			name: "Success case",
			args: ListBucketsArgs{
				ProjectID: projectID,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.ListBucketsFunc = func(ctx context.Context, p string) ([]string, error) {
					return []string{"bucket1", "bucket2"}, nil
				}
			},
			expectErr:      false,
			expectedResult: []string{"bucket1", "bucket2"},
		},
		{
			name: "Fail listing buckets case",
			args: ListBucketsArgs{
				ProjectID: projectID,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.ListBucketsFunc = func(ctx context.Context, p string) ([]string, error) {
					return nil, errors.New("list error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to list buckets: list error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			csMock := &csmocks.MockCloudStorageClient{}
			tc.setupMocks(csMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addListBucketsTool(server, csMock)

			_, res, err := listBucketsToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("listBucketsToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("listBucketsToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}

			if !tc.expectErr {
				resultMap, ok := res.(map[string]any)
				if !ok {
					t.Fatalf("Unexpected result type: %T", res)
				}
				buckets, ok := resultMap["buckets"].([]string)
				if !ok {
					t.Fatalf("Unexpected buckets type: %T", resultMap["buckets"])
				}
				if len(buckets) != len(tc.expectedResult) {
					t.Errorf("Expected result length %d, got %d", len(tc.expectedResult), len(buckets))
				}
			}
		})
	}
}

func TestAddUploadSourceTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	bucketName := "test-bucket"
	destinationDir := "test-dest-dir"

	// Helper function to create a standard temp dir with one file
	createTempDir := func(t *testing.T) (string, func()) {
		tmpDir, err := os.MkdirTemp("", "test-dir-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}

		_, err = os.Create(filepath.Join(tmpDir, "test-file-1.txt"))
		if err != nil {
			t.Fatalf("Failed to create temp file in dir: %v", err)
		}

		cleanup := func() {
			os.RemoveAll(tmpDir)
		}
		return tmpDir, cleanup
	}

	tests := []struct {
		name                   string
		setupFS                func(t *testing.T) (sourcePath string, cleanup func())
		getArgs                func(sourcePath string) UploadSourceArgs
		setupMocks             func(t *testing.T, csMock *csmocks.MockCloudStorageClient)
		expectErr              bool
		expectedErrorSubstring string
		expectedResult string
	}{
		{
			name:    "Success case - bucket exists",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return nil
				}
				csMock.DeleteObjectsFunc = func (ctx context.Context, b string) error {
					return nil
				}
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr: false,
			expectedResult: bucketName,
		},
		{
			name:    "Success case - bucket does not exist",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return storage.ErrBucketNotExist
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return nil
				}
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr: false,
			expectedResult: bucketName,
		},
				{
			name:    "Success case - bucket name not provided",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.GenerateUUIDFunc = func() string {
					return "1"
				}
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return storage.ErrBucketNotExist
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return nil
				}
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr: false,
			expectedResult: projectID + "-1",
		},
		{
			name:    "Fail checking bucket exists case",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return errors.New("check error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to check if bucket exists: check error",
		},
		{
			name:    "Fail creating bucket case",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return storage.ErrBucketNotExist
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return errors.New("create error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to create bucket: create error",
		},
		{
			name:    "Fail accessing source path",
			setupFS: nil,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     "invalid-dir",
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return nil
				}
				csMock.DeleteObjectsFunc = func(ctx context.Context, b string) error {
					return nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to access source path",
		},
		{
			name:    "Fail uploading file case",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return nil
				}
				csMock.DeleteObjectsFunc = func (ctx context.Context, b string) error {
					return nil
				}
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return errors.New("upload error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to upload file: upload error",
		},
		{
			name: "Fail deleted prexisiting objects in bucket",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadSourceArgs {
				return UploadSourceArgs{
					ProjectID:      projectID,
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return nil
				}
				csMock.DeleteObjectsFunc = func (ctx context.Context, b string) error {
					return errors.New("deleting error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to delete objects in bucket: deleting error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sourcePath := ""
			if tc.setupFS != nil {
				var cleanup func()
				sourcePath, cleanup = tc.setupFS(t)
				defer cleanup()
			}
			args := tc.getArgs(sourcePath)

			csMock := &csmocks.MockCloudStorageClient{}
			tc.setupMocks(t, csMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addUploadSourceTool(server, csMock)
			_, res, err := uploadSourceToolFunc(ctx, nil, args)

			if (err != nil) != tc.expectErr {
				t.Errorf("uploadSourceToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("uploadSourceToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}

			if !tc.expectErr {
				resultMap, ok := res.(map[string]any)
				if !ok {
					t.Fatalf("Unexpected result type: %T", res)
				}
				bucketName, ok := resultMap["bucketName"].(string)
				if !ok {
					t.Fatalf("Unexpected type: %T", resultMap["bucketName"])
				}
				if bucketName != tc.expectedResult {
					t.Errorf("Expected result %s, got %s", tc.expectedResult, bucketName)
				}
			}
		})
	}
}
