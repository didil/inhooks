package lib

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewLogger(c *AppConfig) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error
	if c.AppEnv == AppEnvProduction {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize logger")
	}
	return logger, nil
}
