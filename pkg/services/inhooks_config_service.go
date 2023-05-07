package services

import (
	"fmt"
	"os"

	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type InhooksConfigService interface {
	Load(path string) error
	FindFlowForSource(sourceSlug string) *models.Flow
	GetFlow(flowID string) *models.Flow
}

type inhooksConfigService struct {
	logger            *zap.Logger
	inhooksConfig     *models.InhooksConfig
	flowsBySourceSlug map[string]*models.Flow
	flowsByID         map[string]*models.Flow
}

func NewInhooksConfigService(logger *zap.Logger) InhooksConfigService {
	return &inhooksConfigService{
		logger: logger,
	}
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

	err = s.initFlowsMaps()
	if err != nil {
		return errors.Wrapf(err, "failed to build flows map")
	}

	s.log()

	return nil
}

func (s *inhooksConfigService) FindFlowForSource(sourceSlug string) *models.Flow {
	return s.flowsBySourceSlug[sourceSlug]
}

func (s *inhooksConfigService) GetFlow(flowID string) *models.Flow {
	return s.flowsByID[flowID]
}

func (s *inhooksConfigService) initFlowsMaps() error {
	s.flowsBySourceSlug = map[string]*models.Flow{}
	s.flowsByID = map[string]*models.Flow{}
	flowsArr := s.inhooksConfig.Flows

	for _, f := range flowsArr {
		if f.Source == nil {
			return fmt.Errorf("source is empty")
		}
		_, ok := s.flowsBySourceSlug[f.Source.Slug]
		if ok {
			// flow source slug is duplicated
			return fmt.Errorf("flow source slug %s is duplicated", f.Source.Slug)
		}
		s.flowsBySourceSlug[f.Source.Slug] = f

		_, ok = s.flowsByID[f.ID]
		if ok {
			// flow id is duplicated
			return fmt.Errorf("flow id %s is duplicated", f.ID)
		}
		s.flowsByID[f.ID] = f
	}

	return nil
}

func (s *inhooksConfigService) log() {
	for _, f := range s.flowsByID {
		s.logger.Info("loaded flow",
			zap.String("id", f.ID),
			zap.String("sourceID", f.Source.ID),
			zap.String("sourceSlug", f.Source.Slug),
			zap.String("sourceType", string(f.Source.Type)),
		)

		for _, sink := range f.Sinks {
			s.logger.Info("flow sink",
				zap.String("id", sink.ID),
				zap.String("type", string(sink.Type)),
				zap.String("url", string(sink.URL)),
				zap.Duration("delay", sink.Delay),
			)
		}
	}
}
