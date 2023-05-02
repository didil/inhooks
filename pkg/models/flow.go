package models

type Flow struct {
	ID     string  `yaml:"id"`
	Source *Source `yaml:"source"`
	Sinks  []*Sink `yaml:"sinks"`
}
