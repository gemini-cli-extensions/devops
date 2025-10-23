# DevOps MCP Test Client

This directory contains a simple Go-based test client for the `devops-mcp-server`.

It is used to send a test request to a running server instance over HTTP (JSON-RPC).

## 🎯 Purpose

The client is configured to call the `artifactregistry.create_repository` tool with hardcoded arguments. It demonstrates how to:
1.  Connect to the server's HTTP endpoint using the `mark3labs/mcp-go` client library.
2.  Perform the MCP `initialize` handshake.
3.  Make a JSON-RPC tool call (`tools/call`).
4.  Print the result from the server.

## 🚀 How to Use

This client **requires** the `devops-mcp-server` to be built and running in a separate terminal.

### 1. Run the Server

First, build and run the `devops-mcp-server` from its root directory. It *must* be started with the `-http` flag.

```bash
# In terminal 1 (from the server's directory)
# Build the server
go build -o server .

# Run the server on port 8080
./server -http :8080
```

### 2. Run the Client
In a separate terminal, navigate to this testclient directory.

First, ensure the dependencies are downloaded:

```bash
# In terminal 2 (from this directory)
go mod tidy
```
Then, run the client:

```bash
# In terminal 2
go run .
```
The client will connect to the server on http://localhost:8080, initialize, send the tool call, and print the JSON response.
