package services

import (
	"fmt"
	"os"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type InhooksConfigService interface {
	Load(path string) error
	FindFlowForSource(sourceSlug string) *models.Flow
}

type inhooksConfigService struct {
	inhooksConfig     *models.InhooksConfig
	flowsBySourceSlug map[string]*models.Flow
}

func NewInhooksConfigService() InhooksConfigService {
	return &inhooksConfigService{}
}

func (s *inhooksConfigService) Load(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return errors.Wrapf(err, "failed to open inhooks config file")
	}
	defer f.Close()

	inhooksConfig := &models.InhooksConfig{}
	err = yaml.NewDecoder(f).Decode(inhooksConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshall inhooks config file")
	}

	s.inhooksConfig = inhooksConfig

	err = models.ValidateInhooksConfig(inhooksConfig)
	if err != nil {
		return errors.Wrapf(err, "validation err")
	}

	s.flowsBySourceSlug, err = s.buildFlowsMap()
	if err != nil {
		return errors.Wrapf(err, "failed to build flows map")
	}

	return nil
}

func (s *inhooksConfigService) FindFlowForSource(sourceSlug string) *models.Flow {
	return s.flowsBySourceSlug[sourceSlug]
}

func (s *inhooksConfigService) buildFlowsMap() (map[string]*models.Flow, error) {
	flowsMap := map[string]*models.Flow{}
	flowsArr := s.inhooksConfig.Flows

	for _, f := range flowsArr {
		if f.Source == nil {
			return nil, fmt.Errorf("source is empty")
		}
		_, ok := flowsMap[f.Source.Slug]
		if ok {
			// flow id is duplicated
			return nil, fmt.Errorf("flow source slug %s is duplicated", f.Source.Slug)
		}

		flowsMap[f.Source.Slug] = f
	}

	return flowsMap, nil
}
