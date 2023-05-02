package lib

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type AppConfig struct {
	AppEnv AppEnv `env:"APP_ENV"`
	Server *Server
}

type Server struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

func ProcessAppConfig(ctx context.Context) (*AppConfig, error) {
	var appConf *AppConfig
	err := envconfig.Process(ctx, appConf)
	if err != nil {
		return nil, err
	}

	return appConf, nil
}
