# Basic Kubernetes manifests

This document provides examples for generating Kubernetes `Deployment` and `Service` resource manifests that are used with Cloud Deploy to deploy to GKE.

The examples in this file use placeholders between `<` and `>`, e.g. `<DEPLOYMENT_NAME>`. Replace these placeholders with the actual values when generating the YAML.

Cloud Deploy performs build artifact substitutions on the image field so a placeholder like `app-image` is used. When a Cloud Deploy `Release` is created the user supplies the actual image and associates it with the `app-image` key.

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <DEPLOYMENT_NAME>
  labels:
    app: <APPLICATION_NAME>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <APPLICATION_NAME>
  template:
    metadata:
      labels:
        app: <APPLICATION_NAME>
    spec:
      containers:
      - name: <APPLICATION_NAME>
        image: app-image
```

## Kubernetes Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <SERVICE_NAME>
  labels:
    app: <APPLICATION_NAME>
spec:
  selector:
    app: <APPLICATION_NAME>
  ports:
  - protocol: TCP
    port: 80
  type: LoadBalancer
```
