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

package osvclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	scalibr "github.com/google/osv-scalibr"
	scalibrsystem "github.com/google/osv-scalibr/extractor/filesystem/list"
	"github.com/google/osv-scalibr/fs"
	"github.com/google/osv-scalibr/plugin"
)

// contextKey is a private type to use as a key for context values.
type contextKey string

const (
	osvClientKey contextKey = "osvClient"
)

// ClientFrom returns the osvClient stored in the context, if any.
func ClientFrom(ctx context.Context) (OsvClient, bool) {
	client, ok := ctx.Value(osvClientKey).(OsvClient)
	return client, ok
}

// ContextWithClient returns a new context with the provided osvClient.
func ContextWithClient(ctx context.Context, client OsvClient) context.Context {
	return context.WithValue(ctx, osvClientKey, client)
}

// Client is an interface for interacting with the osv API.
type OsvClient interface {
	ScanSecrets(ctx context.Context, root string) (string, error)
}

// clientImpl is a client for interacting with the osv API.
type OsvClientImpl struct {
	osCapabilities *plugin.Capabilities
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (OsvClient, error) {

	capab := &plugin.Capabilities{OS: plugin.OSLinux}
	return &OsvClientImpl{capab}, nil
}

func (o *OsvClientImpl) ScanSecrets(ctx context.Context, root string) (string, error) {
	var allSecretPlugins []plugin.Plugin
	for _, initFns := range scalibrsystem.Secrets {
		for _, initFn := range initFns {
			// initFn() creates the actual extractor instance
			allSecretPlugins = append(allSecretPlugins, initFn())
		}
	}
	// 1. Configure the scan
	cfg := &scalibr.ScanConfig{
		Plugins: allSecretPlugins,
		// Define the scan target: the current directory ("./").
		ScanRoots: fs.RealFSScanRoots(root),

		// Pass the environment capabilities.
		Capabilities: o.osCapabilities,
	}

	// 2. Execute the scan
	log.Println("Starting OSV-SCALIBR scan for secrets...")
	scanner := scalibr.New()
	result := scanner.Scan(ctx, cfg)
	log.Println("Scan complete.")

	// 3. Parse the results
	log.Println("--- OSV-SCALIBR Secret Scan Results ---")

	if len(result.Inventory.Secrets) == 0 {
		log.Println("No inventory (or secrets) found.")
		return "", nil
	}

	// Marshal the findings to JSON for output
	report, err := json.MarshalIndent(result.Inventory.Secrets, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret findings to JSON: %w", err)
	}

	return string(report), nil
}
