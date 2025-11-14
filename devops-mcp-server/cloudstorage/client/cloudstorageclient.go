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

package cloudstorageclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/iam"
	cloudstorage "cloud.google.com/go/storage"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	cloudStorageClientKey contextKey = "cloudStorageClient"
)

// ClientFrom returns the CloudStorageClient stored in the context, if any.
func ClientFrom(ctx context.Context) (CloudStorageClient, bool) {
	client, ok := ctx.Value(cloudStorageClientKey).(CloudStorageClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided CloudStorageClient.
func ContextWithClient(ctx context.Context, client CloudStorageClient) context.Context {
	return context.WithValue(ctx, cloudStorageClientKey, client)
}

// CloudStorageClient is an interface for interacting with the Cloud Storage API.
type CloudStorageClient interface {
	GenerateUUID() string
	// ListBuckets lists the GCS buckets in a specified project.
	ListBuckets(ctx context.Context, projectID string) ([]string, error)
	// CheckBucketExists checks if a GCS bucket exists.
	CheckBucketExists(ctx context.Context, bucketName string) error
	// CreateBucket creates a new GCS bucket.
	CreateBucket(ctx context.Context, projectID, bucketName string) error
	// UploadFile uploads a file to a GCS bucket.
	UploadFile(ctx context.Context, bucketName, objectName string, file *os.File) error
	// CheckObjectExists checks if an object exists in a GCS bucket.
	CheckObjectExists(ctx context.Context, bucketName, objectName string) error
	// GetBucketIamPolicy gets the IAM policy for a GCS bucket.
	GetBucketIamPolicy(ctx context.Context, bucketName string) (*iam.Policy, error)
	// DeleteBucket deletes a GCS bucket.
	DeleteBucket(ctx context.Context, bucketName string) error
	// DeleteObjects deletes all objects from a GCS bucket.
	DeleteObjects(ctx context.Context, bucketName string) error
}

func NewCloudStorageClient(ctx context.Context) (CloudStorageClient, error) {
	c, err := cloudstorage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud storage client: %v", err)
	}
	return &CloudStorageClientImpl{v1client: c}, nil
}


// CloudStorageClientImpl is a client for interacting with the Cloud Storage API.
type CloudStorageClientImpl struct {
	v1client *cloudstorage.Client
}

func (c *CloudStorageClientImpl) GenerateUUID() string {
	return uuid.New().String()
}

// ListBuckets lists the GCS buckets in a specified project.
func (c *CloudStorageClientImpl) ListBuckets(ctx context.Context, projectID string) ([]string, error) {
	var buckets []string
	it := c.v1client.Buckets(ctx, projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, bucketAttrs.Name)
	}
	return buckets, nil
}

// CheckBucketExists checks if a GCS bucket exists.
func (c *CloudStorageClientImpl) CheckBucketExists(ctx context.Context, bucketName string) error {
	// Check if the bucket already exists
	_, err := c.v1client.Bucket(bucketName).Attrs(ctx)
	if err == nil {
		// Bucket already exists, return nil
		return nil
	}
	return err
}

// CreateBucket creates a new GCS bucket.
func (c *CloudStorageClientImpl) CreateBucket(ctx context.Context, projectID, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	bucket := c.v1client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, nil); err != nil {
		return err
	}

	// Make the bucket public by default
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bucket IAM policy: %w", err)
	}
	policy.Add("allUsers", "roles/storage.objectViewer")
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		return fmt.Errorf("failed to set bucket IAM policy: %w", err)
	}

	return nil
}

// UploadFile uploads a file to a GCS bucket.
func (c *CloudStorageClientImpl) UploadFile(ctx context.Context, bucketName, objectName string, file *os.File) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	wc := c.v1client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("failed to copy file to bucket: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// CheckObjectExists checks if an object exists in a GCS bucket.
func (c *CloudStorageClientImpl) CheckObjectExists(ctx context.Context, bucketName, objectName string) error {
	_, err := c.v1client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err == nil {
		return nil
	}
	return err
}

// GetBucketIamPolicy gets the IAM policy for a GCS bucket.
func (c *CloudStorageClientImpl) GetBucketIamPolicy(ctx context.Context, bucketName string) (*iam.Policy, error) {
	policy, err := c.v1client.Bucket(bucketName).IAM().Policy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket IAM policy: %w", err)
	}
	return policy, nil
}

// DeleteBucket deletes a GCS bucket.
func (c *CloudStorageClientImpl) DeleteBucket(ctx context.Context, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := c.v1client.Bucket(bucketName).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}

// DeleteObjects deletes all objects from a GCS bucket.
func (c *CloudStorageClientImpl) DeleteObjects(ctx context.Context, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	it := c.v1client.Bucket(bucketName).Objects(ctx, nil)
	var objectNames []string
	for {
		object, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		objectNames = append(objectNames, object.Name)
	}

	for _, name := range objectNames {
		if err := c.v1client.Bucket(bucketName).Object(name).Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete object %s: %w", name, err)
		}
	}
	return nil
}
