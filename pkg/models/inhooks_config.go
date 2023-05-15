package models

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/didil/inhooks/pkg/lib"
	"golang.org/x/exp/slices"
)

type InhooksConfig struct {
	Flows []*Flow `yaml:"flows"`
}

var idRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{1,255}$`)

func idValidationErr(field string) error {
	return fmt.Errorf("field %s can only contain upper case or lower case letters, digits or hyphens. min length: 1. max length: 255", field)
}

// ValidateInhooksConfig validates inhooks config and sets defaults
func ValidateInhooksConfig(appConf *lib.AppConfig, c *InhooksConfig) error {
	if len(c.Flows) == 0 {
		return fmt.Errorf("no flows defined")
	}

	flowIDs := map[string]bool{}
	sourceSlugs := map[string]bool{}

	for i, f := range c.Flows {
		if !idRegex.MatchString(f.ID) {
			return idValidationErr(fmt.Sprintf("flows[%d].id", i))
		}
		if flowIDs[f.ID] {
			return fmt.Errorf("flow ids must be unique. duplicate flow id: %s", f.ID)
		}
		flowIDs[f.ID] = true

		if f.Source == nil {
			return fmt.Errorf("flow source cannot be empty")
		}

		// validate source
		source := f.Source
		if !idRegex.MatchString(source.ID) {
			return idValidationErr(fmt.Sprintf("flows[%d].source.id", i))
		}

		if !idRegex.MatchString(source.Slug) {
			return idValidationErr(fmt.Sprintf("flows[%d].source.slug", i))
		}

		if sourceSlugs[source.Slug] {
			return fmt.Errorf("flow source slugs must be unique. duplicate source slug: %s", source.Slug)
		}
		sourceSlugs[source.Slug] = true

		if !slices.Contains(SourceTypes, source.Type) {
			return fmt.Errorf("invalid source type: %s. allowed: %v", source.Type, SourceTypes)
		}

		if len(f.Sinks) == 0 {
			return fmt.Errorf("flow sinks cannot be empty")
		}

		for j, sink := range f.Sinks {
			if !idRegex.MatchString(sink.ID) {
				return idValidationErr(fmt.Sprintf("sink[%d].id", j))
			}

			if !slices.Contains(SinkTypes, sink.Type) {
				return fmt.Errorf("invalid sink type: %s. allowed: %v", source.Type, SinkTypes)
			}

			if sink.Delay == nil {
				sink.Delay = &appConf.Sink.DefaultDelay
			}

			if sink.MaxAttempts == nil {
				sink.MaxAttempts = &appConf.Sink.DefaultMaxAttempts
			}

			if sink.RetryExpMultiplier == nil {
				sink.RetryExpMultiplier = &appConf.Sink.DefaultRetryExpMultiplier
			}

			if sink.Type == SinkTypeHttp {
				u, err := url.ParseRequestURI(sink.URL)
				if err != nil {
					return fmt.Errorf("invalid url: %s", sink.URL)
				}
				if u.Scheme != "http" && u.Scheme != "https" {
					return fmt.Errorf("invalid url scheme: %s", sink.URL)
				}
			}
		}
	}

	return nil
}
