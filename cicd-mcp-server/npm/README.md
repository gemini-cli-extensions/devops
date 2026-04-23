# CICD MCP Server NPM Distribution

This directory contains the configuration for distributing the `cicd-mcp-server` Go binary via NPM.

## Overview

The `@google-cloud/cicd-mcp` package is a meta-package that enables users to install and run the `cicd-mcp-server` Go binary across different platforms using standard `npm` or `npx` commands.

### Architecture
It uses a "Meta-package + Platform-specific Sub-packages" pattern:
- **Meta-package (`@google-cloud/cicd-mcp`)**: The main entry point. It contains a `postinstall` script that detects the user's platform and CPU architecture and links the appropriate binary from the installed sub-package.
- **Sub-packages**:
    - `@google-cloud/cicd-mcp-linux-amd64`
    - `@google-cloud/cicd-mcp-darwin-arm64`
    - `@google-cloud/cicd-mcp-windows-amd64`

This approach avoids downloading binaries from GitHub at runtime and relies on NPM's dependency resolution to fetch the correct binary for the user's system.

## GitHub workflows

### Nightly Release (`nightly-release.yml`)
- Runs daily at midnight UTC or on manual trigger.
- Cross-compiles the Go binary for Linux (amd64), Darwin (arm64), and Windows (amd64).
- Creates archives and uploads them to a GitHub release tagged as `nightly`.

### NPM Publish Nightly (`npm-publish-nightly.yml`)
- Can be triggered manually or runs automatically after the `Nightly Release` workflow completes successfully.
- Downloads the latest artifacts from the `nightly` GitHub release.
- Extracts the binaries and places them in the correct sub-package directories.
- Updates the version using a date-based tag (e.g., `X.Y.Z-nightly.YYYYMMDDHHMM`).
- Publishes the packages to NPM with the `--tag nightly` flag.

---

## Manual NPM publishing guide

This section describes the steps to publish the `cicd-mcp` package and its platform-specific sub-packages to the NPM registry manually.

### Prerequisites
- Access to the `@google-cloud` scope on NPM.
- Logged in to NPM via CLI: `npm login`.
- Go installed and configured for cross-compilation.

### Step 1: Build binaries
Ensure all binaries are built and placed in the correct directories. You can use the following commands from the `cicd-mcp-server` directory:

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o npm/cicd-mcp-linux-amd64/bin/cicd-mcp-server .

# Darwin arm64
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o npm/cicd-mcp-darwin-arm64/bin/cicd-mcp-server .

# Windows amd64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o npm/cicd-mcp-windows-amd64/bin/cicd-mcp-server.exe .
```

### Step 2: Publish sub-packages
Navigate to each sub-package directory and publish them. These contain the actual binaries.

```bash
cd npm/cicd-mcp-linux-amd64
npm publish --access public

cd ../cicd-mcp-darwin-arm64
npm publish --access public

cd ../cicd-mcp-windows-amd64
npm publish --access public
```
*Note: Use `--access public` if the `@google-cloud` scope requires it for new packages.*

### Step 3: Publish meta package
Finally, publish the meta package which links everything together.

```bash
cd ../cicd-mcp
npm publish --access public
```

### Automation
TODO: For production releases, these steps should be integrated into the CI/CD pipeline (e.g., GitHub Actions or Cloud Build) triggered on release tags. The nightly process is already automated via the workflows described above.
