package visitor

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func JSONPatch(patch jsonpatch.Patch) Visitor {
	return Func(func(obj *unstructured.Unstructured) error {
		o := obj.UnstructuredContent()
		doc, err := json.Marshal(o)
		if err != nil {
			return err
		}
		doc, err = patch.Apply(doc)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(doc, &o); err != nil {
			return err
		}
		return nil
	})
}
