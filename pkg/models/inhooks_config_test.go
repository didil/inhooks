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
	var hmacAlgorithm HMACAlgorithm = HMACAlgorithmSHA256

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
						Transform: &Transform{
							ID: "js-transform-1",
						},
					},
				},
			},
			{
				ID: "flow-2",
				Source: &Source{
					ID:   "source-2",
					Slug: "source-2-slug",
					Type: "http",
					Verification: &Verification{
						VerificationType:    VerificationTypeHMAC,
						HMACAlgorithm:       &hmacAlgorithm,
						SignatureHeader:     "x-my-header",
						CurrentSecretEnvVar: "FLOW_2_VERIFICATION_SECRET",
					},
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
		TransformDefinitions: []*TransformDefinition{
			{
				ID:     "js-transform-1",
				Type:   TransformTypeJavascript,
				Script: "function transform(data) data.username = data.name end",
			},
		},
	}

	assert.NoError(t, ValidateInhooksConfig(appConf, c))
}

func TestValidateInhooksConfig_NoFlows(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	c := &InhooksConfig{
		Flows: []*Flow{},
	}

	err = ValidateInhooksConfig(appConf, c)
	assert.ErrorContains(t, err, "no flows defined")
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

func TestValidateInhooksConfig_DuplicateFlowIDs(t *testing.T) {
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
						URL:  "https://example.com/sink",
					},
				},
			},
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-2",
					Slug: "source-2-slug",
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

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "flow ids must be unique")
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

func TestValidateInhooksConfig_InvalidVerificationType(t *testing.T) {
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
					Verification: &Verification{
						VerificationType: "random",
					},
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

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid verification type: random. allowed: [hmac]")
}

func TestValidateInhooksConfig_InvalidHMACAlgorithm(t *testing.T) {
	ctx := context.Background()
	appConf, err := testsupport.InitAppConfig(ctx)
	assert.NoError(t, err)

	hmacAlgorithm := HMACAlgorithm("somealgorithm")

	c := &InhooksConfig{
		Flows: []*Flow{
			{
				ID: "flow-1",
				Source: &Source{
					ID:   "source-1",
					Slug: "source-1-slug",
					Type: "http",
					Verification: &Verification{
						VerificationType: VerificationTypeHMAC,
						HMACAlgorithm:    &hmacAlgorithm,
					},
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

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid hmac algorithm: somealgorithm. allowed: [sha256]")
}
func TestValidateInhooksConfig_InexistingTransformID(t *testing.T) {
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
						URL:  "https://example.com/sink",
						Transform: &Transform{
							ID: "non-existent-transform",
						},
					},
				},
			},
		},
		TransformDefinitions: []*TransformDefinition{
			{
				ID:     "js-transform-1",
				Type:   TransformTypeJavascript,
				Script: "function transform(data) end",
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "transform id not found: non-existent-transform")
}

func TestValidateInhooksConfig_InvalidTransformType(t *testing.T) {
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
						URL:  "https://example.com/sink",
						Transform: &Transform{
							ID: "some-transform-1",
						},
					},
				},
			},
		},
		TransformDefinitions: []*TransformDefinition{
			{
				ID:     "some-transform-1",
				Type:   "invalid-type",
				Script: "function transform(data) end",
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "invalid transform type: invalid-type. allowed: [javascript]")
}

func TestValidateInhooksConfig_EmptyTransformScript(t *testing.T) {
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
						URL:  "https://example.com/sink",
						Transform: &Transform{
							ID: "js-transform-1",
						},
					},
				},
			},
		},
		TransformDefinitions: []*TransformDefinition{
			{
				ID:     "js-transform-1",
				Type:   TransformTypeJavascript,
				Script: "",
			},
		},
	}

	assert.ErrorContains(t, ValidateInhooksConfig(appConf, c), "transform script cannot be empty")
}
