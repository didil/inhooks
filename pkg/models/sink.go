package models

type SinkType string

const (
	SinkTypeHttp = "http"
)

var SinkTypes = []SinkType{
	SinkTypeHttp,
}

type Sink struct {
	ID   string   `yaml:"id"`
	Type SinkType `yaml:"type"`
	URL  string   `yaml:"url"`
}
