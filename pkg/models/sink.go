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
	// Process after delay
	Delay *time.Duration `yaml:"delay"`
	// Retry every x time
	RetryInterval *time.Duration `yaml:"retryInterval"`
	// Retry exponential multiplier. 1 is constant backoff. Set to > 1 for exponential backoff.
	RetryExpMultiplier *float64 `yaml:"retryExpMultiplier"`
	// Max attempts
	MaxAttempts *int `yaml:"maxAttempts"`
}
