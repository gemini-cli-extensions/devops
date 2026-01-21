# Local RAG

This module is used to build a local RAG database using [chromem-go](https://pkg.go.dev/github.com/clocklear/chromem-go#section-readme) in-memory vector store.

## Pre-requisites

- Google Cloud project with billing enabled.
- Google Cloud project with VertexAI APIs enabled.
- Application Default Credentials (ADC) for a Google Cloud account. With default project, or a quota-project enabled. e.g., via 
  ```
  gcloud auth application-default login
  gcloud auth application-default set-quota-project PROJECT_ID
  ```
