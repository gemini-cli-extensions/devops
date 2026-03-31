# How to Deploy to Google Kubernetes Engine (GKE) with `kubectl`

This document outlines the standard procedure for deploying containerized applications to a Google Kubernetes Engine (GKE) cluster using standard CLI tools.

---

## Step 1: Gather Deployment Parameters

Before executing any commands, the following parameters must be identified:

*   **`PROJECT_ID`**: The Google Cloud project where the GKE cluster and Artifact Registry are hosted.
*   **`CLUSTER_NAME`**: The name of the GKE cluster.
*   **`LOCATION`**: The geographic location (zone or region) of the cluster (e.g., `us-central1-a` or `us-central1`).
*   **`IMAGE_URI`**: The full URI of the container image in Artifact Registry (e.g., `us-central1-docker.pkg.dev/my-project/my-repo/my-image:latest`).
*   **`MANIFEST_PATH`**: The path to the Kubernetes manifest files (e.g., `./k8s/`).

---

## Step 2: Authenticate with the GKE Cluster

Use the `gcloud` CLI to retrieve the necessary credentials for the cluster. This configures `kubectl` to point to the correct GKE instance.

```bash
gcloud container clusters get-credentials <CLUSTER_NAME> \
    --location=<LOCATION> \
    --project=<PROJECT_ID>
```

---

## Step 3: Prepare Kubernetes Manifests

Ensure that the Kubernetes manifests (typically `deployment.yaml` and `service.yaml`) are correctly configured.

### 1. Update the Image URI
The `image:` field in the `Deployment` manifest MUST be updated to point to the `IMAGE_URI` generated during the build step.

### 2. Standard Manifest Example (if none exist)
If the project does not already have manifests, generate a basic set:

**`deployment.yaml`**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <APP_NAME>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <APP_NAME>
  template:
    metadata:
      labels:
        app: <APP_NAME>
    spec:
      containers:
      - name: <APP_NAME>
        image: <IMAGE_URI>
        ports:
        - containerPort: <CONTAINER_PORT>
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

**`service.yaml`**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: <APP_NAME>-service
spec:
  selector:
    app: <APP_NAME>
  ports:
    - protocol: TCP
      port: 80
      targetPort: <CONTAINER_PORT>
  type: LoadBalancer
```

---

## Step 4: Deploy Build Artifacts

Use the `kubectl apply` command to deploy the manifests to the cluster.

```bash
kubectl apply -f <MANIFEST_PATH>
```

---

## Step 5: Verification

After the deployment is applied, provide the user with commands to check the status of the rollout and retrieve the external IP address (if using a LoadBalancer).

1.  **Check Pod Status**: `kubectl get pods`
2.  **Check Service IP**: `kubectl get service <APP_NAME>-service`
3.  **Check Rollout Status**: `kubectl rollout status deployment/<APP_NAME>`
