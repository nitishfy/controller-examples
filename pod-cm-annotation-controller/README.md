# Pod-ConfigMap Annotation Sync Controller

A lightweight Kubernetes controller built with `controller-runtime` that automatically syncs data from a specified ConfigMap into a Pod's annotations.

## Overview

This controller watches:
- **Pods** as the primary resource
- **ConfigMaps** as secondary resources

If a Pod has an annotation in the form:

```yaml
metadata:
  annotations:
    configmap-name: sample-config
```

Then the controller will fetch the sample-config ConfigMap and inject its data into the Pod's annotations like:

```
metadata:
  annotations:
    data.key1: value1
    data.key2: value2
```

If the ConfigMap has no data, the Pod will be annotated with:
```
data: empty
```

and all other keys that the controller has added(starting with `data.`) will be removed.

## Installation

1. Clone this repository:
   ```bash
   git clone https://github.com/nitishfy/controller-examples.git
   ```
2. Build and run the controller locally (ensure the kubeconfig points to a running cluster)
   ```bash
   go run main.go
   ```