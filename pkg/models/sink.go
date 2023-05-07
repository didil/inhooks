package models

import "time"

type SinkType string

const (
	SinkTypeHttp = "http"
)

var SinkTypes = []SinkType{
	SinkTypeHttp,
}

type Sink struct {
	// Sink ID
	ID string `yaml:"id"`
	// Sink Type
	Type SinkType `yaml:"type"`
	// Sink Url for HTTP sinks
	URL string `yaml:"url"`
	// Deliver after delay
	Delay time.Duration `yaml:"delay"`
}
