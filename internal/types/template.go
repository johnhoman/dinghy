package types

type TemplateConfig struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Templates  []string       `json:"templates" yaml:"templates"`
	Values     map[string]any `json:"values" yaml:"values"`
}
