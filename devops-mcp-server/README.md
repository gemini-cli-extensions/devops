# MCP Server

This server provides a local interface to interact with Google Cloud services through the Gemini CLI.

## Capabilities

The Google Cloud DevOps MCP Server provides a suite of tools for:

*   **Artifact Registry:** Setting up and managing repositories.
*   **Cloud Build:** Managing and running build triggers.
*   **Cloud Run:** Deploying and managing services from images or source.
*   **Cloud Storage:** Managing buckets and uploading source code.
*   **Developer Connect:** Establishing connections to git repositories.
*   **Security:** Scanning for secrets using OSV.
*   **Search/RAG:** Querying common CI/CD patterns and knowledge snippets.

For a full list of available tools and their arguments, see the [Tool Reference](REFERENCE.md).

## Prerequisites

*   **Google Cloud SDK:** Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) and authenticate with your Google account.
*   **Set up Application Default Credentials (ADC):**
    ```bash
    gcloud auth application-default login
    ```

## Running the Server

### 1. Using stdio (Recommended for local use)

This method is recommended for local development and integration with Gemini CLI.

**Build the server:**

```bash
go build -o devops-mcp-server
```

**Gemini CLI Configuration:**

Update your Gemini CLI `settings.json` to include the following:

```json
{
  "mcpServers": {
    "devops": {
      "command": "/path/to/devops-mcp-server",
      "cwd": "/path/to/server-directory"
    }
  }
}
```

Replace `/path/to/devops-mcp-server` with the absolute path to the built binary and `/path/to/server-directory` with the project root.

### 2. Using streamable HTTP

You can also run the server as an HTTP service.

**Run the server:**

```bash
go run . -http :9000
```

**Gemini CLI Configuration:**

Update your Gemini CLI `settings.json` to include the following:

```json
{
  "mcpServers": {
    "devops": {
      "httpUrl": "http://localhost:9000/mcp"
    }
  }
}
```
