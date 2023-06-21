package mutate

import (
	"context"
	"fmt"
	"github.com/johnhoman/dinghy/internal/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
)

var (
	_ Mutator = &ImageResolver{}
)

// ImageResolver resolves image tags for workload resources
// to the digest. It will pull your local docker configuration
// for access to any private registries
type ImageResolver struct{}

func (i *ImageResolver) Name() string {
	return "builtin.dinghy.dev/imageTagResolver"
}

func (i *ImageResolver) Visit(obj *resource.Object) error {
	switch obj.GroupVersionKind().GroupKind() {
	case schema.GroupKind{Group: "", Kind: "Pod"}:
		pod := &corev1.Pod{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pod); err != nil {
			return err
		}
		for k := range pod.Spec.Containers {
			image := pod.Spec.Containers[k].Image
			if image == "" {
				continue
			}

		}
	case schema.GroupKind{Group: "apps", Kind: "ReplicaSet"}:
		panic("not implemented")
	case schema.GroupKind{Group: "apps", Kind: "StatefulSet"}:
		panic("not implemented")
	case schema.GroupKind{Group: "apps", Kind: "Deployment"}:
		panic("not implemented")
	case schema.GroupKind{Group: "apps", Kind: "DaemonSet"}:
		panic("not implemented")
	case schema.GroupKind{Group: "batch", Kind: "Job"}:
		panic("not implemented")
	case schema.GroupKind{Group: "batch", Kind: "CronJob"}:
		panic("not implemented")
	}
	return nil
}

func getImageDigest(imageName, tag string) (digest string, err error) {
	// Create an HTTP client
	client := http.DefaultClient

	// Send a HEAD request to get the image manifest
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", imageName, tag)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		err = resp.Body.Close()
	}()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to retrieve image manifest: %s", resp.Status)
	}

	// Get the Docker-Content-Digest header
	digest = resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", fmt.Errorf("image digest not found in response headers")
	}

	return digest, nil
}
