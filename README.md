# Dinghy
A new kustomize that's extensible

## Goals
1. Add additional config fields to the kustomization yaml
2. Improve speed (caching instead of temporary directories)

```yaml
apiVersion: dinghy.dev/v1beta1
kind: Config
```