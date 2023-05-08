package testsupport

import (
	"context"

	"github.com/didil/inhooks/pkg/lib"
)

func InitAppConfig(ctx context.Context) (*lib.AppConfig, error) {
	err := lib.LoadEnvFromFile("../../.env.test")
	if err != nil {
		return nil, err
	}

	appConf, err := lib.InitAppConfig(ctx)
	if err != nil {
		return nil, err
	}

	return appConf, nil
}
