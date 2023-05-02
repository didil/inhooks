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
	GetFlow(id string) *models.Flow
}

type inhooksConfigService struct {
	inhooksConfig *models.InhooksConfig
	flows         map[string]*models.Flow
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

	s.flows, err = s.buildFlowsMap()
	if err != nil {
		return errors.Wrapf(err, "failed to build flows map")
	}

	return nil
}

func (s *inhooksConfigService) GetFlow(id string) *models.Flow {
	return s.flows[id]
}

func (s *inhooksConfigService) buildFlowsMap() (map[string]*models.Flow, error) {
	flowsMap := map[string]*models.Flow{}
	flowsArr := s.inhooksConfig.Flows

	for _, f := range flowsArr {
		_, ok := flowsMap[f.ID]
		if ok {
			// flow id is duplicated
			return nil, fmt.Errorf("flow id %s is duplicated", f.ID)
		}

		flowsMap[f.ID] = f
	}

	return flowsMap, nil
}
