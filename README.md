# deployvia

_Was my application successfully deployed?_

## Description

**deployvia** is an API that allows you to check the status of your applications deployed to Atlas by Argo CD.

## Development

**TODO**: Automate the process below.

### How to release new version

1. Increment the version in `VERSION` file.

1. Update `images.newTag` in `manifests/base/kustomization.yaml` to the new version.

1. Run this command:

```bash
kustomize build manifests/base > manifests/install.yaml
```

1. Commit and push the changes.

1. Create a new release on GitHub with the tag `v<version>`, e.g. `v0.1.0`.
