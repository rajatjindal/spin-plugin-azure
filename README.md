# Spin Azure Plugin

A CLI tool for deploying and managing [Spin](https://github.com/fermyon/spin) applications on Azure Kubernetes Service (AKS) with workload identity.

## Installation

```bash
make build
```

## Prerequisites

- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) (`az` command)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Spin CLI](https://github.com/fermyon/spin)
- [Helm](https://helm.sh/docs/intro/install/) - package manager for Kubernetes
- An Azure subscription

## Usage

### Login to Azure

```bash
spin-azure login
```

### Create a new AKS cluster

```bash
spin-azure cluster create --name my-cluster --resource-group my-rg --location eastus
```

This creates a complete environment in one command:
- AKS cluster with workload identity enabled
- Spin Operator installed
- Azure managed identity called "workload-identity"
- Kubernetes service account configured for the identity

You can specify a custom identity name:

```bash
spin-azure cluster create --name my-cluster --resource-group my-rg --create-identity=my-custom-identity
```

### Use an existing AKS cluster

```bash
spin-azure cluster use --name existing-cluster --resource-group existing-rg
```

When using an existing cluster, you can optionally install the Spin Operator:

```bash
spin-azure cluster use --name existing-cluster --resource-group existing-rg --install-spin-operator
```

And you can create an identity at the same time:

```bash
spin-azure cluster use --name existing-cluster --resource-group existing-rg --create-identity=my-custom-identity
```

### Check workload identity status

```bash
spin-azure cluster check-identity
```

This checks if workload identity is enabled on the current cluster, and enables it if not.

### Install Spin Operator

You can install the Spin Operator on an existing cluster:

```bash
spin-azure cluster install-spin-operator
```

### Bind to Azure CosmosDB

```bash
spin-azure bind cosmosdb --name my-cosmos --resource-group my-rg
```

This grants your workload identity access to the specified CosmosDB instance.

### Deploy a Spin application

You can deploy a Spin application to your cluster with a simple command:

```bash
spin-azure deploy --from path/to/spinapp.yaml
```

This will deploy the application using the default "workload-identity" that was created with the cluster.

If you want to use a custom identity, you can specify it with the --identity flag:

```bash
spin-azure deploy --from path/to/spinapp.yaml --identity my-custom-identity
```

> warning: since SpinApp CRD does not support serviceAccountName yet, you need to edit the deployment YAML file to set the `serviceAccountName` field to `workload-identity`.

