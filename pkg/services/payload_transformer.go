package services

import (
	"context"

	"github.com/didil/inhooks/pkg/models"
)

type PayloadTransformer interface {
	Transform(ctx context.Context, transformDefinition *models.TransformDefinition, payload []byte) ([]byte, error)
}

type payloadTransformer struct {
}

func NewPayloadTransformer() PayloadTransformer {
	return &payloadTransformer{}
}

func (p *payloadTransformer) Transform(ctx context.Context, transformDefinition *models.TransformDefinition, payload []byte) ([]byte, error) {
	return payload, nil
}
