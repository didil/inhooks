package lib

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type AppConfig struct {
	AppEnv            AppEnv `env:"APP_ENV"`
	InhooksConfigFile string `env:"INHOOKS_CONFIG_FILE,default=inhooks.yml"`
	Server            ServerConfig
	Redis             RedisConfig
	Supervisor        SupervisorConfig
	HTTPClient        HTTPClientConfig
	Sink              SinkConfig
}

type ServerConfig struct {
	Host                string        `env:"HOST"`
	Port                int           `env:"PORT,default=3000"`
	ShutdownGracePeriod time.Duration `env:"SERVER_SHUTDOWN_GRACE_PERIOD,default=5s"`
}

type RedisConfig struct {
	URL           string `env:"REDIS_URL,default=redis://localhost:6379"`
	InhooksDBName string `env:"REDIS_INHOOKS_DB_NAME"`
}

// Supervisor queues handling settings
type SupervisorConfig struct {
	ReadyWaitTime     time.Duration `env:"SUPERVISOR_READY_WAIT_TIME,default=5s"`
	ErrSleepTime      time.Duration `env:"SUPERVISOR_ERR_SLEEP_TIME,default=5s"`
	SchedulerInterval time.Duration `env:"SUPERVISOR_SCHEDULER_INTERVAL,default=30s"`
}

type HTTPClientConfig struct {
	Timeout time.Duration `env:"HTTP_CLIENT_TIMEOUT,default=5s"`
}

type SinkConfig struct {
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
