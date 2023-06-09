package kustomize

type mutatePlugin struct {
	mutator   Mutator
	newConfig func() any
}

var (
	mutateRegistry = make(map[string]*mutatePlugin)
)

func RegisterMutate(name string, mutator Mutator, newConfig func() any) {
	if newConfig == nil {
		newConfig = func() any {
			return make(map[string]any)
		}
	}
	mutateRegistry[name] = &mutatePlugin{
		mutator:   mutator,
		newConfig: newConfig,
	}
}

func isRegisteredMutator(name string) bool {
	_, ok := mutateRegistry[name]
	return ok
}

func getRegisteredMutator(name string) *mutatePlugin {
	m, ok := mutateRegistry[name]
	if !ok {
		panic("check for registered mutator before getting it: use isRegisteredMutator")
	}
	return m
}
