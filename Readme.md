# Artifactory provisioning operator

Creates Artifactory Docker registries and the users to access them based on instances of the Project CRD of [Kubi](https://github.com/ca-gip/kubi).

In an `intranet` cluster, for the following Project resource :
```yaml
---
apiVersion: "ca-gip.github.com/v1"
kind: Project
metadata:
  name: sampleorg-project1-development
spec:
    tenant: sampleorg
    environment: development
    project: project1
    stages:
      - scratch
      - staging
      - stable
```

The operator will :
- Create Artifactory Docker registries for the three stages: `sampleorg-project1-docker-scratch-intranet`, etc.
- Create a `reader` and `writer` user in Artifactory with permissions on the above registries.
- Create a `dl_sampleorg_project1` group in Artifactory with RW rights on the above registries.
- Create a [Docker registry secret](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-in-the-cluster-that-holds-your-authorization-token) named `project-registries` in the `sampleorg-projectA-development` namespace (same name as the project).

## Build Requirements
- Go 1.10 minimum (tested with v1.10.4).
- [Glide](https://github.com/Masterminds/glide).
- GNU Make.
- Docker to build the image.

## Runtime requirements
- Administrative access to a Kubernetes cluster with [Kubi](https://github.com/ca-gip/kubi) deployed and its Project CRD installed.
- Administrative access to an activated Artifactory Pro installation, with a well-configured [reverse-proxy for Docker Subdomains](https://www.jfrog.com/confluence/display/RTF/Configuring+a+Reverse+Proxy#ConfiguringaReverseProxy-UsingSubdomain).

## Building locally

```bash
make dep
make build
```
The output binary will be called `build/artifactory-operator`.

## Testing locally
After building, edit `tests/operator-env.inc.sh` and set valid values for your environment (see below for configuration settings). The operator will use the current context of your `.kube/config` file to connect to the K8S cluster.

Then :
```bash
source tests/operator-env.inc.sh
build/artifactory-operator
```

## Building the Docker image
Edit the `DOCKER_REPO` variable in `Makefile` to target a Docker registry where you can push the image, then:
```bash
make image
```

## Deploying to Kubernetes
Edit `deployment/01 - artifactory-operator-config.yaml` and `deployment/02 - artifactory-operator-secret.yaml` to adapt the configuration (see below for options).
Edit `deployment/03 - artifactory-operator.yaml` to pull the image from your own Docker registry.

Then :
```bash
kubectl apply -f deployment/*
```