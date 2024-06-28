package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMessageTransformer_Transform_Javascript(t *testing.T) {
	config := &lib.TransformConfig{
		JavascriptTimeout: 5000 * time.Millisecond,
	}

	mt := NewMessageTransformer(config)

	m := &models.Message{
		Payload: []byte(`{
				"name": "John Doe",
				"age": 30,
				"locations": ["New York", "London", "Tokyo"],
				"scores": [85, 90, 78, 92]
			}`),
		HttpHeaders: http.Header{
			"Content-Type":  []string{"application/json"},
			"X-Request-Id":  []string{"123"},
			"Authorization": []string{"Bearer token123"},
		},
	}
	transformDefinition := &models.TransformDefinition{
		Type: models.TransformTypeJavascript,
		Script: `
			function transform(bodyStr, headers) {
				const body = JSON.parse(bodyStr);
				body.username = body.name;
				delete body.name;
				delete body.age;
				body.location_count = body.locations.length;
				body.average_score = body.scores.reduce((a, b) => a + b, 0) / body.scores.length;
				headers["X-AUTH-TOKEN"] = [headers.Authorization[0].split(' ')[1]];
				delete headers.Authorization;
				return [JSON.stringify(body), headers];
			}
			`,
	}
	err := mt.Transform(context.Background(), transformDefinition, m)
	assert.NoError(t, err)

	assert.JSONEq(t, `{"username":"John Doe","locations":["New York","London","Tokyo"],"scores":[85,90,78,92],"average_score":86.25,"location_count":3}`, string(m.Payload))
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}, "X-AUTH-TOKEN": []string{"token123"}, "X-Request-Id": []string{"123"}}, m.HttpHeaders)
}

func TestMessageTransformer_Transform_Javascript_Error(t *testing.T) {
	config := &lib.TransformConfig{
		JavascriptTimeout: 5000 * time.Millisecond,
	}

	mt := NewMessageTransformer(config)

	m := &models.Message{
		Payload: []byte(`{
				"name": "John Doe",
				"age": 30,
				"locations": ["New York", "London", "Tokyo"],
				"scores": [85, 90, 78, 92]
			}`),
		HttpHeaders: http.Header{
			"Content-Type":  []string{"application/json"},
			"X-Request-Id":  []string{"123"},
			"Authorization": []string{"Bearer token123"},
		},
	}
	transformDefinition := &models.TransformDefinition{
		Type: models.TransformTypeJavascript,
		Script: `
			function transform(bodyStr, headers) {
				const body = JSON.parse(bodyStr);
				throw new Error("random error while in the transform function");
				return [JSON.stringify(body), headers];
			}
			`,
	}
	err := mt.Transform(context.Background(), transformDefinition, m)
	assert.ErrorContains(t, err, "failed to transform message: failed to execute JavaScript: Error: random error while in the transform function")
}
