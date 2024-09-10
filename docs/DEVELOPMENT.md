# Development Guide

This document provides guidelines and instructions for setting up the development environment and contributing to the project.

## Set Environment Variables

Create `.env` file in the root directory of the project and add the following environment variables:

```sh
QUAY_AUTH_TOKEN=<your-quay-auth-token>
QUAY_USER_NAME=<your-quay-username>
```

## Viewing available Make targets

To view all available Make targets, run:

```sh
make help
```

## Building Operator image

To build and push the operator image to Quay.io, run:

```sh
make container-build
make container-push
```

## Running Tests

You can run the unit tests using the following command:

```sh
make test
```

## Running Lint

To run the lint checks, execute:

```sh
make lint
```

## Deploying Operator

Ensure KUBECONFIG points to target OpenShift cluster. To deploy the operator, run:

```sh
make deploy
```

## Create `ConsoleApplication` Custom Resource (CR)

To create a `ConsoleApplication` CR, run:

```sh
kubectl create -k examples/success.yaml
```

## Uninstalling Operator

Ensure KUBECONFIG points to target OpenShift cluster. Let's begin by deleting the payload image first with:

```sh
kubectl delete -k examples/success.yaml
```

Subsequently, proceed with uninstalling the operator:

```sh
make undeploy
```
