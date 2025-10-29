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

func TestAddCreateBucketTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	bucketName := "test-bucket"

	tests := []struct {
		name          string
		args          CreateBucketArgs
		setupMocks    func(*csmocks.MockCloudStorageClient)
		expectErr     bool
		expectedErrorSubstring string
	}{
		{
			name: "Success case - bucket already exists",
			args: CreateBucketArgs{
				ProjectID:  projectID,
				BucketName: bucketName,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return nil
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return nil
				}
			},
			expectErr: false,
		},
		{
			name: "Success case - bucket created",
			args: CreateBucketArgs{
				ProjectID:  projectID,
				BucketName: bucketName,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return storage.ErrBucketNotExist
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return nil
				}
			},
			expectErr: false,
		},
		{
			name: "Fail creating bucket case",
			args: CreateBucketArgs{
				ProjectID:  projectID,
				BucketName: bucketName,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return storage.ErrBucketNotExist
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return errors.New("creation failed")
				}
			},
			expectErr:     true,
			expectedErrorSubstring: "failed to create bucket: creation failed",
		},
		{
			name: "Failed to check bucket exists case",
			args: CreateBucketArgs{
				ProjectID:  projectID,
				BucketName: bucketName,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.CheckBucketExistsFunc = func(ctx context.Context, b string) error {
					return errors.New("attrs error")
				}
				csMock.CreateBucketFunc = func(ctx context.Context, p, b string) error {
					return nil
				}
			},
			expectErr:     true,
			expectedErrorSubstring: "failed to check if bucket exists: attrs error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			csMock := &csmocks.MockCloudStorageClient{}
			tc.setupMocks(csMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addCreateBucketTool(server, csMock)

			_, _, err := createBucketToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("createBucketToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

            if tc.expectErr {
                if err == nil {
                    t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
                } else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
                    t.Errorf("createBucketToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
                }
            }
		})
	}
}
func TestAddUploadFileTool(t *testing.T) {
	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "test-object"

    // Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up the file afterwards

	filePath := tmpfile.Name()

	tests := []struct {
		name          string
		args          UploadFileArgs
		setupMocks    func(*csmocks.MockCloudStorageClient)
		expectErr     bool
		expectedErrorSubstring string
	}{
		{
			name: "Success case",
			args: UploadFileArgs{
				BucketName: bucketName,
				ObjectName: objectName,
				FilePath:   filePath,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr: false,
		},
		{
			name: "Fail opening source file case",
			args: UploadFileArgs{
				BucketName: bucketName,
				ObjectName: objectName,
				FilePath:   "invalid-file.txt",
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr:     true,
			expectedErrorSubstring: "failed to open source file",
		},
        {
			name: "Fail uploading file case",
			args: UploadFileArgs{
				BucketName: bucketName,
				ObjectName: objectName,
				FilePath:   filePath,
			},
			setupMocks: func(csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return errors.New("upload error")
				}
			},
			expectErr:     true,
			expectedErrorSubstring: "failed to upload file: upload error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			csMock := &csmocks.MockCloudStorageClient{}
			tc.setupMocks(csMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addUploadFileTool(server, csMock)

			_, _, err := uploadFileToolFunc(ctx, nil, tc.args)

			if (err != nil) != tc.expectErr {
				t.Errorf("uploadFileToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
                if err == nil {
                    t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
                } else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
                    t.Errorf("uploadFileToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
                }
            }
		})
	}
}

func TestAddUploadDirectoryTool(t *testing.T) {
	ctx := context.Background()
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
		getArgs                func(sourcePath string) UploadDirectoryArgs
		setupMocks             func(t *testing.T, csMock *csmocks.MockCloudStorageClient)
		expectErr              bool
		expectedErrorSubstring string
	}{
		{
			name:    "Success case",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadDirectoryArgs {
				return UploadDirectoryArgs{
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil // Success
				}
			},
			expectErr: false,
		},
		{
			name:    "Fail accessing source path (dir does not exist)",
			setupFS: nil,
			getArgs: func(sourcePath string) UploadDirectoryArgs {
				return UploadDirectoryArgs{
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     "invalid-dir",
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return nil
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to access source path",
		},
		{
			name:    "Fail upload file case",
			setupFS: createTempDir,
			getArgs: func(sourcePath string) UploadDirectoryArgs {
				return UploadDirectoryArgs{
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					return errors.New("upload error")
				}
			},
			expectErr:              true,
			expectedErrorSubstring: "failed to upload file: upload error",
		},
		{
			name: "Success with nested directory",
			setupFS: func(t *testing.T) (string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}

				// Create root file
				_, err = os.Create(filepath.Join(tmpDir, "root-file.txt"))
				if err != nil {
					t.Fatalf("Failed to create root file: %v", err)
				}

				// Create nested directory and file
				nestedDir := filepath.Join(tmpDir, "nested")
				if err := os.Mkdir(nestedDir, 0755); err != nil {
					t.Fatalf("Failed to create nested dir: %v", err)
				}
				_, err = os.Create(filepath.Join(nestedDir, "nested-file.txt"))
				if err != nil {
					t.Fatalf("Failed to create nested file: %v", err)
				}

				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			getArgs: func(sourcePath string) UploadDirectoryArgs {
				return UploadDirectoryArgs{
					BucketName:     bucketName,
					DestinationDir: destinationDir,
					SourcePath:     sourcePath,
				}
			},
			setupMocks: func(t *testing.T, csMock *csmocks.MockCloudStorageClient) {
				// We expect two uploads with specific object names
				expectedObjects := map[string]bool{
					"test-dest-dir/root-file.txt":     false,
					"test-dest-dir/nested/nested-file.txt": false,
				}

				csMock.UploadFileFunc = func(ctx context.Context, b, o string, f *os.File) error {
					if _, ok := expectedObjects[o]; !ok {
						t.Errorf("Uploaded unexpected object: %s", o)
					}
					expectedObjects[o] = true // Mark as seen
					return nil // Success
				}
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup Filesystem
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
			addUploadDirectoryTool(server, csMock)
			_, _, err := uploadDirectoryToolFunc(ctx, nil, args)

			if (err != nil) != tc.expectErr {
				t.Errorf("uploadDirectoryToolFunc() error = %v, expectErr %v", err, tc.expectErr)
			}

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tc.expectedErrorSubstring)
				} else if !strings.Contains(err.Error(), tc.expectedErrorSubstring) {
					t.Errorf("uploadDirectoryToolFunc() error = %q, expectedErrorSubstring %q", err.Error(), tc.expectedErrorSubstring)
				}
			}
		})
	}
}