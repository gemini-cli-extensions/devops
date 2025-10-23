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

package clouddeploy

import (
	"context"
	"fmt"

	deploy "cloud.google.com/go/deploy/apiv1"
	"google.golang.org/api/iterator"
	deploypb "cloud.google.com/go/deploy/apiv1/deploypb"
)

// Client is a client for interacting with the Cloud Deploy API.
type Client struct {
	client *deploy.CloudDeployClient
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	c, err := deploy.NewCloudDeployClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud deploy client: %v", err)
	}
	return &Client{client: c}, nil
}

// CreateDeliveryPipeline creates a new Cloud Deploy delivery pipeline.
func (c *Client) CreateDeliveryPipeline(ctx context.Context, projectID, location, pipelineID, description string) (*deploypb.DeliveryPipeline, error) {
	req := &deploypb.CreateDeliveryPipelineRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		DeliveryPipelineId: pipelineID,
		DeliveryPipeline: &deploypb.DeliveryPipeline{
			Description: description,
			Pipeline: &deploypb.DeliveryPipeline_SerialPipeline{
				SerialPipeline: &deploypb.SerialPipeline{},
			},
		},
	}
	op, err := c.client.CreateDeliveryPipeline(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery pipeline: %v", err)
	}
	return op.Wait(ctx)
}

// CreateGKETarget creates a new Cloud Deploy GKE target.
func (c *Client) CreateGKETarget(ctx context.Context, projectID, location, targetID, gkeCluster, description string) (*deploypb.Target, error) {
	req := &deploypb.CreateTargetRequest{
		Parent:   fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		TargetId: targetID,
		Target: &deploypb.Target{
			Description: description,
		},
	}
	op, err := c.client.CreateTarget(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create gke target: %v", err)
	}
	return op.Wait(ctx)
}

// CreateCloudRunTarget creates a new Cloud Deploy Cloud Run target.
func (c *Client) CreateCloudRunTarget(ctx context.Context, projectID, location, targetID, description string) (*deploypb.Target, error) {
	req := &deploypb.CreateTargetRequest{
		Parent:   fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		TargetId: targetID,
		Target: &deploypb.Target{
			Description: description,
		},
	}
	op, err := c.client.CreateTarget(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run target: %v", err)
	}
	return op.Wait(ctx)
}

// CreateRollout creates a new Cloud Deploy rollout.
func (c *Client) CreateRollout(ctx context.Context, projectID, location, pipelineID, releaseID, rolloutID, targetID string) (*deploypb.Rollout, error) {
	req := &deploypb.CreateRolloutRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", projectID, location, pipelineID, releaseID),
		RolloutId: rolloutID,
		Rollout: &deploypb.Rollout{
			TargetId: targetID,
		},
	}
	op, err := c.client.CreateRollout(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollout: %v", err)
	}
	return op.Wait(ctx)
}

// ListDeliveryPipelines lists all Cloud Deploy delivery pipelines.
func (c *Client) ListDeliveryPipelines(ctx context.Context, projectID, location string) ([]*deploypb.DeliveryPipeline, error) {
	req := &deploypb.ListDeliveryPipelinesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	}
	it := c.client.ListDeliveryPipelines(ctx, req)
	var pipelines []*deploypb.DeliveryPipeline
	for {
		pipeline, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list delivery pipelines: %v", err)
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, nil
}

// ListTargets lists all Cloud Deploy targets.
func (c *Client) ListTargets(ctx context.Context, projectID, location string) ([]*deploypb.Target, error) {
	req := &deploypb.ListTargetsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	}
	it := c.client.ListTargets(ctx, req)
	var targets []*deploypb.Target
	for {
		target, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list targets: %v", err)
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// ListReleases lists all Cloud Deploy releases for a given delivery pipeline.
func (c *Client) ListReleases(ctx context.Context, projectID, location, pipelineID string) ([]*deploypb.Release, error) {
	req := &deploypb.ListReleasesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", projectID, location, pipelineID),
	}
	it := c.client.ListReleases(ctx, req)
	var releases []*deploypb.Release
	for {
		release, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list releases: %v", err)
		}
		releases = append(releases, release)
	}
	return releases, nil
}

// ListRollouts lists all Cloud Deploy rollouts for a given release.
func (c *Client) ListRollouts(ctx context.Context, projectID, location, pipelineID, releaseID string) ([]*deploypb.Rollout, error) {
	req := &deploypb.ListRolloutsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", projectID, location, pipelineID, releaseID),
	}
	it := c.client.ListRollouts(ctx, req)
	var rollouts []*deploypb.Rollout
	for {
		rollout, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list rollouts: %v", err)
		}
		rollouts = append(rollouts, rollout)
	}
	return rollouts, nil
}