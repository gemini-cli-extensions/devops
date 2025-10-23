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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Client is a client for interacting with Google Cloud Storage.
type Client struct {
	client *storage.Client
	projectID string
}

// NewClient creates a new Client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}
	return &Client{
		client:    client,
		projectID: projectID,
	}, nil
}

// CreateBucket creates a new GCS bucket.
func (c *Client) CreateBucket(ctx context.Context, projectID, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Check if the bucket already exists
	_, err := c.client.Bucket(bucketName).Attrs(ctx)
	if err == nil {
		// Bucket already exists, return nil
		return nil
	}
	if err != storage.ErrBucketNotExist {
		// An unexpected error occurred while checking for the bucket
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	// Bucket does not exist, proceed to create it
	err = c.client.Bucket(bucketName).Create(ctx, projectID, nil)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// UploadFile uploads a file to a GCS bucket.
func (c *Client) UploadFile(ctx context.Context, projectID, bucketName, objectName, filePath string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	wc := c.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("failed to copy file to bucket: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// ReadFile reads a file from a GCS bucket.
func (c *Client) ReadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	r, err := c.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// UploadDirectory uploads a directory to a GCS bucket.
func (c *Client) UploadDirectory(ctx context.Context, projectID, bucketName, destinationDir, sourcePath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        err = func() error {
            relPath, err := filepath.Rel(sourcePath, path)
            if err != nil {
                return fmt.Errorf("failed to get relative path: %w", err)
            }

            objectName := filepath.Join(destinationDir, relPath)
            // Ensure objectName uses forward slashes for GCS compatibility
            objectName = strings.ReplaceAll(objectName, "\\", "/")

            file, err := os.Open(path)
            if err != nil {
                return fmt.Errorf("failed to open file %s: %w", path, err)
            }
            defer file.Close() // This defer is now scoped to this anonymous function

            wc := c.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
            if _, err := io.Copy(wc, file); err != nil {
                return fmt.Errorf("failed to copy file %s to bucket: %w", path, err)
            }
            if err := wc.Close(); err != nil {
                return fmt.Errorf("failed to close writer for file %s: %w", path, err)
            }
            return nil
        }()

        return err // Return the error from the anonymous function to filepath.Walk
    })
}
