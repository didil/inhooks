package services

import (
	"fmt"
	"os"

	"github.com/didil/inhooks/pkg/lib"
	"github.com/didil/inhooks/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type InhooksConfigService interface {
	Load(path string) error
	FindFlowForSource(sourceSlug string) *models.Flow
	GetFlow(flowID string) *models.Flow
	GetFlows() map[string]*models.Flow
	GetTransformDefinition(transformID string) *models.TransformDefinition
}

type inhooksConfigService struct {
	logger                   *zap.Logger
	appConf                  *lib.AppConfig
	inhooksConfig            *models.InhooksConfig
	flowsBySourceSlug        map[string]*models.Flow
	flowsByID                map[string]*models.Flow
	transformDefinitionsByID map[string]*models.TransformDefinition
}

func NewInhooksConfigService(logger *zap.Logger, appConf *lib.AppConfig) InhooksConfigService {
	return &inhooksConfigService{
		logger:  logger,
		appConf: appConf,
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

	// set defaults

	err = models.ValidateInhooksConfig(s.appConf, inhooksConfig)
	if err != nil {
		return errors.Wrapf(err, "validation err")
	}

	err = s.initFlowsMaps()
	if err != nil {
		return errors.Wrapf(err, "failed to build flows map")
	}

	err = s.initTransformDefinitionsMap()
	if err != nil {
		return errors.Wrapf(err, "failed to build transform definitions map")
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

func (s *inhooksConfigService) GetFlows() map[string]*models.Flow {
	return s.flowsByID
}

func (s *inhooksConfigService) GetTransformDefinition(transformID string) *models.TransformDefinition {
	return s.transformDefinitionsByID[transformID]
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

func (s *inhooksConfigService) initTransformDefinitionsMap() error {
	s.transformDefinitionsByID = map[string]*models.TransformDefinition{}
	for _, transformDefinition := range s.inhooksConfig.TransformDefinitions {
		s.transformDefinitionsByID[transformDefinition.ID] = transformDefinition
	}

	return nil
}

func (s *inhooksConfigService) log() {
	for _, transform := range s.inhooksConfig.TransformDefinitions {
		s.logger.Info("loaded transform",
			zap.String("id", transform.ID),
			zap.String("type", string(transform.Type)),
		)
	}

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
				zap.Durationp("delay", sink.Delay),
				zap.Any("transform", sink.Transform),
			)
		}
	}
}
