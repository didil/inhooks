package lib

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type AppConfig struct {
	AppEnv AppEnv `env:"APP_ENV"`
	Server Server
	Redis  Redis
}

type Server struct {
	Host string `env:"HOST,default=localhost"`
	Port int    `env:"PORT,default=3000"`
}

type Redis struct {
	URL           string `env:"REDIS_URL,default=redis://localhost:6379"`
	InhooksDBName string `env:"REDIS_INHOOKS_DB_NAME"`
}

func ProcessAppConfig(ctx context.Context) (*AppConfig, error) {
	appConf := &AppConfig{}
	err := envconfig.Process(ctx, appConf)
	if err != nil {
		return nil, err
	}

	return appConf, nil
}
