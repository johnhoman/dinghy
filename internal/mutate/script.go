package mutate

import "github.com/johnhoman/dinghy/internal/visitor"

type scriptConfig struct {
	Script string         `yaml:"script" dinghy:"required"`
	Config map[string]any `yaml:"config" dinghy:"required"`
}

// Script runs a javascript snippet, passing is the current resource
// for mutation. The script MUST define a single function `mutate`
// with the signature mutate(obj, config), which wil be called with the
// resource being visited as well as the config provided to the mutator
func Script(config any) (visitor.Visitor, error) {
	c, ok := config.(*scriptConfig)
	if !ok {
		return nil, ErrTypedConfig
	}
	return visitor.Script(c.Script, c.Config)
}
