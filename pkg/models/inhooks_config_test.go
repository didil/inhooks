package models

import (
	"context"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestValidateInhooksConfig_OK(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	delay := 12 * time.Minute
	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:    "sink-1",
						Type:  "http",
						URL:   "https://example.com/sink",
						Delay: &delay,
					},
				},
			},
			{
				ID: "flow-2",
				Source: &Source{
					ID:   "source-2",
					Slug: "source-2-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "http",
						URL:  "https://example.com/sink",
					},
				},
			},
		},
	}

	assert.NoError(t, ValidateInhooksConfig(appConf, c))
}

func TestValidateInhooksConfig_MissingID(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "http",
						URL:  "https://example.com/sink",
					},
				},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "field flows[0].id can only contain upper case or lower case letters, digits or hyphens. min length: 1. max length: 255")
}

func TestValidateInhooksConfig_DuplicateSlug(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "http",
						URL:  "https://example.com/sink",
					},
				},
			},
			{
				ID: "flow-2",
				Source: &Source{
					ID:   "source-2",
					Slug: "source-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-2",
						Type: "http",
						URL:  "https://example.com/sink",
					},
				},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "flow source slugs must be unique")
}

func TestValidateInhooksConfig_EmptySinks(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
				},
				Sinks: []*Sink{},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "flow sinks cannot be empty")
}

func TestValidateInhooksConfig_InvalidSourceType(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "abc",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "http",
						URL:  "https://example.com/sink",
					},
				},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid source type: abc. allowed: [http]")
}

func TestValidateInhooksConfig_InvalidSinkType(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "xyz",
						URL:  "https://example.com/sink",
					},
				},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid sink type: http. allowed: [http]")
}

func TestValidateInhooksConfig_InvalidSinkUrl(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
				},
				Sinks: []*Sink{
					{
						ID:   "sink-1",
						Type: "http",
						URL:  "ABCD123",
					},
				},
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid url: ABCD123")
}
