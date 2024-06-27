package models

type TransformType string

const (
	TransformTypeLua = "lua"
)

var TransformTypes = []TransformType{
	TransformTypeLua,
}

type TransformDefinition struct {
	ID     string        `yaml:"id"`
	Type   TransformType `yaml:"type"`
	Script string        `yaml:"script"`
}

type Transform struct {
	ID string `yaml:"id"`
}
