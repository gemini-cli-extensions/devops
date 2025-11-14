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

	addListBucketsTool(server, c)
	addUploadSourceTool(server, c)
	return nil
}

type ListBucketsArgs struct {
	ProjectID string `json:"project_id" jsonschema:"The Google Cloud project ID."`
}

var listBucketsToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args ListBucketsArgs) (*mcp.CallToolResult, any, error)

func addListBucketsTool(server *mcp.Server, csClient cloudstorageclient.CloudStorageClient) {
	listBucketsToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args ListBucketsArgs) (*mcp.CallToolResult, any, error) {
		res, err := csClient.ListBuckets(ctx, args.ProjectID)
		if err != nil {
			return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to list buckets: %w", err)
		}
		return &mcp.CallToolResult{}, map[string]any{"buckets": res}, nil

	}
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.list_buckets", Description: "Lists Cloud Storage buckets in a specified project."}, listBucketsToolFunc)
}

type UploadSourceArgs struct {
	ProjectID      string `json:"project_id" jsonschema:"The Google Cloud project ID."`
	BucketName     string `json:"bucket_name,omitempty" jsonschema:"The name of the bucket. Optional."`
	DestinationDir string `json:"destination_dir" jsonschema:"The name of the destination directory."`
	SourcePath     string `json:"source_path" jsonschema:"The path to the source directory."`
}

var uploadSourceToolFunc func(ctx context.Context, req *mcp.CallToolRequest, args UploadSourceArgs) (*mcp.CallToolResult, any, error)

func addUploadSourceTool(server *mcp.Server, csClient cloudstorageclient.CloudStorageClient) {
	uploadSourceToolFunc = func(ctx context.Context, req *mcp.CallToolRequest, args UploadSourceArgs) (*mcp.CallToolResult, any, error) {
		if args.BucketName == "" {
			args.BucketName = fmt.Sprintf("%s-%s", args.ProjectID, csClient.GenerateUUID())
		}

		// Check if bucket exists, and create bucket if it does not.
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
		} else {
			// Delete all existing objects in bucket
			if err := csClient.DeleteObjects(ctx, args.BucketName); err != nil {
				return &mcp.CallToolResult{}, nil, fmt.Errorf("failed to delete objects in bucket: %w", err)
			}
		}

		// Upload all files in source path to destination directory in bucket.
		return &mcp.CallToolResult{}, map[string]any{"bucketName": args.BucketName}, filepath.Walk(args.SourcePath, func(path string, info os.FileInfo, err error) error {
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
	mcp.AddTool(server, &mcp.Tool{Name: "cloudstorage.upload_source", Description: "Uploads source to a GCS bucket. If a new bucket is created, it will create a public bucket."}, uploadSourceToolFunc)
}
