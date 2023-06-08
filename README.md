# Kustomize
A new kustomize that's extensible

## Goals
1. Add additional config fields to the kustomization yaml
2. Improve speed (caching instead of temporary directories)

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
```

## Modules

```go
package main
import (
	"github.com/johnhoman/kustomize"
)

func main() {
	kustomize.RegisterModule(nil)
	if err := cmd.Run(); err != nil {
		panic(err)
    }
}
```