#!/usr/bin/env bash
#
# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail


readonly MCP_SERVER_DIR="devops-mcp-server"
readonly BINARY_NAME="devops-mcp-server"


# Prints an error message to standard error.
err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z') ERROR]: $*" >&2
}

# Checks if the 'go' is available.
is_go_installed() {

  INSTALLED_GO_VERSION=$(apt show golang-go 2>/dev/null | grep Version | awk '{print $2}' | head -n 1)

  if [[ -z "${INSTALLED_GO_VERSION}" ]]; then
    err "golang-go package does not seem to be installed."
    exit 1
  fi

  REQUIRED_GO_VERSION="1.24.0"
  if dpkg --compare-versions "${INSTALLED_GO_VERSION}" "lt" "${REQUIRED_GO_VERSION}"; then
    err "golang-go ${INSTALLED_GO_VERSION} is older than ${REQUIRED_GO_VERSION}."
    exit 1
  fi
}


# Installs Go using apt.
install_go() {
  echo "Go tooling not found. Installing golang-go..."
  # Recommended method for gLinux: go/go-getting-started
  sudo apt update && sudo apt install -y golang-go || { err "apt install failed."; return 1; }
  echo "Go installed successfully."

  # Attempt to refresh the current shell environment.
  if [[ -f "${HOME}/.bashrc" ]]; then
    echo "Sourcing ${HOME}/.bashrc to update PATH in current session..."
    # shellcheck source=/dev/null
    source "${HOME}/.bashrc"
  fi
}

# Builds the MCP Server
build_mcp() {
  echo "Building '${MCP_SERVER_DIR}'..."
  if [[ ! -d "${MCP_SERVER_DIR}" ]]; then
    err "Error: MCP serverdirectory '${MCP_SERVER_DIR}' not found."
    return 1
  fi
  (
    cd ${MCP_SERVER_DIR}
    go mod tidy
    go build -o "../${BINARY_NAME}"
    echo "Successfully built '${BINARY_NAME}' Please move it to `.gemini/extensions/devops/bin/devops-mcp-server`"
  )
}

main() {
  if ! is_go_installed; then
    echo "Go tooling not detected."
    install_go || { err "Go installation failed. Exiting."; exit 1; }

    # Re-check if 'go' is available after installation and sourcing.
    if ! is_go_installed; then
      err "Go was installed, but still does not meet minimum Go tooling needed."
      exit 1
    fi
    echo "Go is now detected."
  fi

  echo ""
  echo "Building MCP Server"
  build_mcp || { err "Go build for MCP server failed. Exiting."; exit 1; }
}

main "$@"
