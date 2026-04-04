package models

type TransformType string

const (
	TransformTypeJavascript = "javascript"
)

var TransformTypes = []TransformType{
	TransformTypeJavascript,
}

type TransformDefinition struct {
	ID         string        `yaml:"id"`
	Type       TransformType `yaml:"type"`
	Script     string        `yaml:"script"`
	ScriptPath string        `yaml:"script_path"`
}

type Transform struct {
	ID string `yaml:"id"`
}
