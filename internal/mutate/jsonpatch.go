package mutate

import (
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/johnhoman/dinghy/internal/visitor"
)

type JSONPatch jsonpatch.Patch

func (j *JSONPatch) Visit(obj *unstructured.Unstructured) error {
	return visitor.JSONPatch(jsonpatch.Patch(*j)).Visit(obj)
}
