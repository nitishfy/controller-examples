# 🔐 Secret Synchronizer Controller

This is a Kubernetes controller built using [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) that keeps a central secret in sync across multiple namespaces based on a specific label.

## ✨ Features

- Watches a `Secret` named `central-secret` in the `default` namespace.
- Automatically copies or updates the secret in all namespaces labeled with `secret-sync=true`.
- If the central secret is deleted, it cleans up the synced secrets from all target namespaces.

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Kubernetes cluster (e.g., Kind, Minikube)
- `kubectl`
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)

### Installation

1. **Clone the repository**  
   ```bash
   git clone https://github.com/your-username/secret-synchronizer-controller.git
   cd secret-synchronizer-controller

2. # 🔐 Secret Synchronizer Controller

This is a Kubernetes controller built using [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) that keeps a central secret in sync across multiple namespaces based on a specific label.

## ✨ Features

- Watches a `Secret` named `central-secret` in the `default` namespace.
- Automatically copies or updates the secret in all namespaces labeled with `secret-sync=true`.
- If the central secret is deleted, it cleans up the synced secrets from all target namespaces.

## 📦 Project Structure

- `main.go`: Contains the main controller logic and reconciliation loop.
- Modular methods for:
  - Fetching the central secret
  - Listing eligible namespaces
  - Creating or updating secrets
  - Cleaning up when the central secret is deleted

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Kubernetes cluster (e.g., Kind, Minikube)
- `kubectl`
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)

### Installation

1. **Clone the repository**  
   ```bash
   git clone https://github.com/your-username/secret-synchronizer-controller.git
   cd secret-synchronizer-controller

2. **Apply the central secret**
   ```
   # secret.yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: central-secret
      namespace: default
    type: Opaque
    data:
      username: dXNlcg==
      password: cGFzc3dvcmQ=
  ```

3. **Label your target namespace(s)**
```
kubectl label ns team-a secret-sync=true
```

4. **Run the controller locally**
```
go run main.go
```
