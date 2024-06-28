package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/dop251/goja"
)

type MessageTransformer interface {
	Transform(ctx context.Context, transformDefinition *models.TransformDefinition, m *models.Message) error
}

type messageTransformer struct {
	config *lib.TransformConfig
}

func NewMessageTransformer(config *lib.TransformConfig) MessageTransformer {
	return &messageTransformer{
		config: config,
	}
}

func (mt *messageTransformer) Transform(ctx context.Context, transformDefinition *models.TransformDefinition, m *models.Message) error {
	switch transformDefinition.Type {
	case models.TransformTypeJavascript:
		transformedPayload, transformedHeaders, err := mt.runJavascriptTransform(m.Payload, m.HttpHeaders, transformDefinition.Script)
		if err != nil {
			return fmt.Errorf("failed to transform message: %w", err)
		}
		m.Payload = transformedPayload
		m.HttpHeaders = transformedHeaders
		return nil

	default:
		return fmt.Errorf("unknown transform type: %s", transformDefinition.Type)
	}
}

func (mt *messageTransformer) runJavascriptTransform(payload []byte, headers http.Header, jsScript string) ([]byte, http.Header, error) {
	vm := goja.New()

	time.AfterFunc(mt.config.JavascriptTimeout, func() {
		vm.Interrupt("halt")
	})

	vm.Set("bodyStr", string(payload))

	headersStr, err := json.Marshal(headers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal headers to JSON: %w", err)
	}
	vm.Set("headersStr", string(headersStr))

	// Prepare the full script
	fullScript := fmt.Sprintf(`
		/* User Function */
		%s
		/* End User Function */

		const headers = JSON.parse(headersStr);
		var results = transform(bodyStr, headers);
		[results[0], results[1]];
	`, jsScript)

	// Run the script
	val, err := vm.RunString(fullScript)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute JavaScript: %w", err)
	}

	// Get the results
	results := val.Export().([]interface{})
	if len(results) != 2 {
		return nil, nil, fmt.Errorf("expected 2 results in js transform, got %d", len(results))
	}
	transformedPayloadStr, ok := results[0].(string)
	if !ok {
		return nil, nil, fmt.Errorf("expected payload to be of type string, got %T", results[0])
	}
	transformedHeadersTemp, ok := results[1].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("expected headers to be of type map[string]interface{}, got %T", results[1])
	}

	// build back the header object
	transformedHeaders := http.Header{}
	for k, values := range transformedHeadersTemp {
		valuesArr, ok := values.([]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("expected header values to be of type []string, got %T", values)
		}

		stringValuesArr := make([]string, len(valuesArr))
		for i, value := range valuesArr {
			stringValue, ok := value.(string)
			if !ok {
				return nil, nil, fmt.Errorf("expected header value to be of type string, got %T", value)
			}
			stringValuesArr[i] = stringValue
		}

		transformedHeaders[k] = stringValuesArr
	}

	return []byte(transformedPayloadStr), transformedHeaders, nil
}
