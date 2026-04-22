# @google-cloud/cicd-mcp

This is the main package for the CI/CD Model Context Protocol (MCP) Server.

## Usage

You can run the server directly using `npx`:

```bash
npx @google-cloud/cicd-mcp
```

## Installation

If you prefer to install it as a dependency:

```bash
npm install @google-cloud/cicd-mcp
```

## How it works

This package acts as a wrapper. Upon installation, it detects your operating system and architecture and attempts to use the appropriate platform-specific binary provided by one of the optional dependencies:

- `@google-cloud/cicd-mcp-linux-amd64`
- `@google-cloud/cicd-mcp-darwin-arm64`
- `@google-cloud/cicd-mcp-windows-amd64`

## License

Apache-2.0
