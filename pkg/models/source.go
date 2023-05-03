package models

type SourceType string

const (
	SourceTypeHttp = "http"
)

var SourceTypes = []SourceType{
	SourceTypeHttp,
}

type Source struct {
	ID   string     `yaml:"id"`
	Slug string     `yaml:"slug"`
	Type SourceType `yaml:"type"`
}
