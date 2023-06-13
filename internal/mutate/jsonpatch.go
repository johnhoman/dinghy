package mutate

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/johnhoman/dinghy/internal/visitor"
)

func JSONPatch(config any) (visitor.Visitor, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	patch := jsonpatch.Patch{}
	if err := json.Unmarshal(raw, &patch); err != nil {
		return nil, err
	}
	return visitor.JSONPatch(patch), nil
}
