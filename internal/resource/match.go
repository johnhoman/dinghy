package resource

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// MatchKinds limits a match query to just the provided kinds.
// MatchKinds doesn't replace the existing kinds on matchOptions,
// so multiple MatchKinds options can be used together.
// Map.Matches(MatchKinds("Pod"), MatchKinds("StatefulSet"))
func MatchKinds(kinds ...schema.GroupVersionKind) MatchOption {
	return func(o *matchOptions) {
		if len(kinds) == 0 {
			return
		}
		if o.kinds == nil {
			o.kinds = sets.New[schema.GroupVersionKind]()
		}
		o.kinds.Insert(kinds...)
	}
}

func MatchNames(names ...string) MatchOption {
	return func(o *matchOptions) {
		if len(names) == 0 {
			return
		}
		if o.names == nil {
			o.names = sets.New[string]()
		}
		o.names.Insert(names...)
	}
}

func MatchNamespaces(namespaces ...string) MatchOption {
	return func(o *matchOptions) {
		if len(namespaces) == 0 {
			return
		}
		if o.namespaces == nil {
			o.namespaces = sets.New[string]()
		}
		o.names.Insert(namespaces...)
	}
}

func MatchLabels(labels map[string]string) MatchOption {
	return func(o *matchOptions) {
		if len(labels) == 0 {
			return
		}
		if o.matchLabels == nil {
			o.matchLabels = make(map[string]string)
		}
		for key, value := range labels {
			o.matchLabels[key] = value
		}
	}
}

func MatchAnnotations(annotations map[string]string) MatchOption {
	return func(o *matchOptions) {
		if len(annotations) == 0 {
			return
		}
		if o.matchAnnotations == nil {
			o.matchAnnotations = make(map[string]string)
		}
		for key, value := range annotations {
			o.matchLabels[key] = value
		}
	}
}

// A MatchOption is used to limit the set of resources
// returned by a Map.Matches query.
type MatchOption func(o *matchOptions)

type matchOptions struct {
	kinds            sets.Set[schema.GroupVersionKind]
	names            sets.Set[string]
	namespaces       sets.Set[string]
	matchLabels      map[string]string
	matchAnnotations map[string]string
}
