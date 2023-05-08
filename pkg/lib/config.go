package lib

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type AppConfig struct {
	AppEnv     AppEnv `env:"APP_ENV"`
	Server     Server
	Redis      Redis
	Supervisor Supervisor
	HTTPClient HTTPClient
	Sink       Sink
}

type Server struct {
	Host                string        `env:"HOST,default=localhost"`
	Port                int           `env:"PORT,default=3000"`
	ShutdownGracePeriod time.Duration `env:"SERVER_SHUTDOWN_GRACE_PERIOD,default=5s"`
}

type Redis struct {
	URL           string `env:"REDIS_URL,default=redis://localhost:6379"`
	InhooksDBName string `env:"REDIS_INHOOKS_DB_NAME"`
}

// Supervisor queues handling settings
type Supervisor struct {
	ReadyWaitTime time.Duration `env:"SUPERVISOR_READY_WAIT_TIME,default=5s"`
}

type HTTPClient struct {
	Timeout time.Duration `env:"HTTP_CLIENT_TIMEOUT,default=5s"`
}

type Sink struct {
	DefaultDelay       time.Duration `env:"SINK_DEFAULT_DELAY,default=0"`
	DefaultMaxAttempts int           `env:"SINK_DEFAULT_MAX_ATTEMPTS,default=3"`
	DefaultRetryAfter  time.Duration `env:"SINK_DEFAULT_RETRY_AFTER,default=0"`
}

func InitAppConfig(ctx context.Context) (*AppConfig, error) {
	appConf := &AppConfig{}
	err := envconfig.Process(ctx, appConf)
	if err != nil {
		return nil, err
	}

	return appConf, nil
}
