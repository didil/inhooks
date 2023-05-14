package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger_Dev(t *testing.T) {
	conf := &AppConfig{
		AppEnv: AppEnvDevelopment,
	}
	logger, err := NewLogger(conf)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("testing dev logger")
}

func TestNewLogger_Prod(t *testing.T) {
	conf := &AppConfig{
		AppEnv: AppEnvProduction,
	}
	logger, err := NewLogger(conf)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("testing prod logger")
}
