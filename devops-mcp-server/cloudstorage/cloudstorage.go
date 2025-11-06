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

package cloudstorage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	cloudstorageclient "devops-mcp-server/cloudstorage/client"

	cloudstorage "cloud.google.com/go/storage"
)

// AddTools adds all cloud storage related tools to the mcp server.
// It expects the cloudstorageclient and mcp.Server to be in the context.
func AddTools(ctx context.Context, server *mcp.Server) error {
	c, ok := cloudstorageclient.ClientFrom(ctx)
	if !ok {
		return fmt.Errorf("cloud storage client not found in context")
	}

	addCreateBucketTool(server, c)
	addUploadFileTool(server, c)
	addUploadDirectoryTool(server, c)
	return nil
}

type CreateBucketArgs struct {
	ProjectID  string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	BucketName string `json:"bucket_name" jsonschema:"The name of the bucket."`
}

var createBucketToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args CreateBucketArgs) (*mcp.CallToolResult, any, error)

func addCreateBucketTool(server *mcp.Server, csClient cloudstorageclient.CloudStorageClient) {
	createBucketToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args CreateBucketArgs) (*mcp.CallToolResult, any, error) {
		err := csClient.CheckBucketExists(ctx, args.BucketName)
		if err != nil {
			if !errors.Is(err, cloudstorage.ErrBucketNotExist) {
				// An unexpected error occurred while checking for the bucket
				return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to check if bucket exists: %w", err)
			}
			err = csClient.CreateBucket(ctx, args.ProjectID, args.BucketName)
			if err != nil {
				return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to create bucket: %w", err)
			}
		}
		return &mcp.CallToolResult{}, nil, nil
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.create_bucket", Description: "Creates a new Cloud Storage bucket."}, createBucketToolFunc)
}

type UploadFileArgs struct {
	BucketName string `json:"bucket_name" jsonschema:"The name of the bucket."`
	ObjectName string `json:"object_name" jsonschema:"The name of the object to upload to the bucket."`
	FilePath   string `json:"file_path" jsonschema:"The path to the source file."`
}

var uploadFileToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args UploadFileArgs) (*mcp.CallToolResult, any, error)

func addUploadFileTool(server *mcp.Server, csClient cloudstorageclient.CloudStorageClient) {
	uploadFileToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args UploadFileArgs) (*mcp.CallToolResult, any, error) {
		file, err := os.Open(args.FilePath)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to open source file: %w", err)
		}
		defer file.Close()

		err = csClient.UploadFile(ctx, args.BucketName, args.ObjectName, file)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to upload file: %w", err)
		}
		return &mcp.CallToolResult{}, nil, nil

	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.upload_file", Description: "Uploads a file to a Cloud Storage bucket."}, uploadFileToolFunc)
}

type UploadDirectoryArgs struct {
	BucketName     string `json:"bucket_name" jsonschema:"The name of the bucket."`
	DestinationDir string `json:"destination_dir" jsonschema:"The name of the destination directory."`
	SourcePath     string `json:"source_path" jsonschema:"The path to the source directory."`
}

var uploadDirectoryToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args UploadDirectoryArgs) (*mcp.CallToolResult, any, error)

func addUploadDirectoryTool(server *mcp.Server, csClient cloudstorageclient.CloudStorageClient) {
	uploadDirectoryToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args UploadDirectoryArgs) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{}, nil, filepath.Walk(args.SourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("failed to access source path: %w", err)
			}

			if info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(args.SourcePath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			objectName := filepath.Join(args.DestinationDir, relPath)
			// Ensure objectName uses forward slashes for GCS compatibility
			objectName = strings.ReplaceAll(objectName, "\\", "/")

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close() // This defer is now scoped to this anonymous function

			err = csClient.UploadFile(ctx, args.BucketName, objectName, file)
			if err != nil {
				return fmt.Errorf("failed to upload file: %w", err)
			}
			return nil
		})
	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.upload_directory", Description: "Uploads a directory to a Cloud Storage bucket."}, uploadDirectoryToolFunc)
}
