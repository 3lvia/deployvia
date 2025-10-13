# deployvia

_Was my application successfully deployed?_ 🚀

**deployvia** is an API that allows you to check the status of your applications deployed to Atlas by Argo CD.
Its purpose is to bridge the gap between traditional push-based CI/CD and pull-based GitOps CD.

### Reasoning

At Elvia, we are used to deploying an application and waiting until it is successfully deployed before moving on the the next step in the pipeline.
With Argo CD, the deployment process is asynchronous, and it can be hard to know when the application is ready.
This is where **deployvia** comes in. It allows you to check the status of your Argo CD deployment inside your pipeline.

### Overview

- Runs in the same Kubernetes cluster and namespace as an Argo CD instance and watches `Application` resources with specific labels.
- Provides a single endpoint (`/deployment`) for checking the status of a deployment:
  - endpoint requires a valid GitHub Actions OIDC token from the 3lvia organization.
  - endpoint is protected by Traefik (reverse proxy) using IP whitelisting and rate limiting. Only Elvia-hosted GitHub runners can access it.
  - endpoint will not respond to the request until the matching `Application` has finished syncing and the correct image tag is deployed.

## Development

### Running locally

```
make run
```

### Monitoring

We have two Grafana dashboards [here](https://elvia.grafana.net/dashboards/f/eepxnbvcdmbcwc/argocd).

### How to release new version

**TODO**: Automate this process more.

1. Increment the version in `VERSION` file.

2. Update `images.newTag` in `manifests/base/kustomization.yaml` to the new version.

3. Run this command:

```bash
kustomize build manifests/base > manifests/install.yaml
```

4. Commit and push the changes.
