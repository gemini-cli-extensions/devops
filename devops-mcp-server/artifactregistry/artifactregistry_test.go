package artifactregistry

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	armocks "devops-mcp-server/artifactregistry/client/mocks"
	iammocks "devops-mcp-server/iam/client/mocks"

	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	cloudresourcemanagerv1 "google.golang.org/api/cloudresourcemanager/v1"
)

func TestAddSetupRepositoryTool(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	location := "us-central1"
	repositoryID := "test-repo"
	format := "DOCKER"
	saEmail := "test-sa@example.com"

	repo := &artifactregistrypb.Repository{
		Name:   fmt.Sprintf("projects/%s/locations/%s/repositories/%s", projectID, location, repositoryID),
		Format: artifactregistrypb.Repository_DOCKER,
	}

	tests := []struct {
		name          string
		args          SetupRepoArgs
		setupMocks    func(*armocks.MockArtifactRegistryClient, *iammocks.MockIAMClient)
		expectErr     bool
		expectedError string
	}{
		{
			name: "Success case",
			args: SetupRepoArgs{
				ProjectID:           projectID,
				Location:            location,
				RepositoryID:        repositoryID,
				Format:              format,
				ServiceAccountEmail: saEmail,
			},
			setupMocks: func(arMock *armocks.MockArtifactRegistryClient, iamMock *iammocks.MockIAMClient) {
				arMock.CreateRepositoryFunc = func(ctx context.Context, p, l, r, f string) (*artifactregistrypb.Repository, error) {
					return repo, nil
				}
				iamMock.EXPECT().AddIAMRoleBinding(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&cloudresourcemanagerv1.Policy{}, nil)
			},
			expectErr: false,
		},
		{
			name: "Already Exists case",
			args: SetupRepoArgs{
				ProjectID:    projectID,
				Location:     location,
				RepositoryID: repositoryID,
				Format:       format,
			},
			setupMocks: func(arMock *armocks.MockArtifactRegistryClient, iamMock *iammocks.MockIAMClient) {
				arMock.CreateRepositoryFunc = func(ctx context.Context, p, l, r, f string) (*artifactregistrypb.Repository, error) {
					return nil, errors.New("rpc error: code = AlreadyExists desc = repository already exists")
				}
				arMock.GetRepositoryFunc = func(ctx context.Context, p, l, r string) (*artifactregistrypb.Repository, error) {
					return repo, nil
				}
			},
			expectErr: false,
		},
		{
			name: "IAM Fails case",
			args: SetupRepoArgs{
				ProjectID:           projectID,
				Location:            location,
				RepositoryID:        repositoryID,
				Format:              format,
				ServiceAccountEmail: saEmail,
			},
			setupMocks: func(arMock *armocks.MockArtifactRegistryClient, iamMock *iammocks.MockIAMClient) {
				arMock.CreateRepositoryFunc = func(ctx context.Context, p, l, r, f string) (*artifactregistrypb.Repository, error) {
					return repo, nil
				}
				iamMock.EXPECT().AddIAMRoleBinding(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("iam failed"))
			},
			expectErr:     true,
			expectedError: "repository created, but failed to grant permissions: iam failed",
		},
		{
			name: "Create Fails case",
			args: SetupRepoArgs{
				ProjectID:    projectID,
				Location:     location,
				RepositoryID: repositoryID,
				Format:       format,
			},
			setupMocks: func(arMock *armocks.MockArtifactRegistryClient, iamMock *iammocks.MockIAMClient) {
				arMock.CreateRepositoryFunc = func(ctx context.Context, p, l, r, f string) (*artifactregistrypb.Repository, error) {
					return nil, errors.New("creation failed")
				}
			},
			expectErr:     true,
			expectedError: "failed to create repository: creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			arMock := &armocks.MockArtifactRegistryClient{}
			iamMock := iammocks.NewMockIAMClient(ctrl)
			tt.setupMocks(arMock, iamMock)

			server := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{})
			addSetupRepositoryTool(server, arMock, iamMock)

			_, _, err := setupRepositoryToolFunc(ctx, nil, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("setupRepositoryToolFunc() error = %v, expectErr %v", err, tt.expectErr)
			}

			if tt.expectErr && err.Error() != tt.expectedError {
				t.Errorf("setupRepositoryToolFunc() error = %q, expectedError %q", err.Error(), tt.expectedError)
			}
		})
	}
}
