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
	ReadyWaitTime time.Duration `env:"SUPERVISOR_READY_WAIT_TIME,default=5s"`
	ErrSleepTime  time.Duration `env:"SUPERVISOR_ERR_SLEEP_TIME,default=5s"`
	// interval between scheduler runs to move scheduled jobs to "ready for processing" queue
	SchedulerInterval time.Duration `env:"SUPERVISOR_SCHEDULER_INTERVAL,default=30s"`
	// interval to move back stuck messages from processing to ready queue
	ProcessingRecoveryInterval time.Duration `env:"SUPERVISOR_PROCESSING_RECOVERY_INTERVAL,default=5m"`
	// enables deleting done messages from the database after DoneQueueCleanupDelay
	DoneQueueCleanupEnabled bool `env:"SUPERVISOR_DONE_QUEUE_CLEANUP_ENABLED,default=false"`
	// delay after which done messages are deleted from the database. Default 14 days = 336 hours
	DoneQueueCleanupDelay time.Duration `env:"SUPERVISOR_DONE_QUEUE_CLEANUP_DELAY,default=336h"`
	// interval between done queue cleanup runs
	DoneQueueCleanupInterval time.Duration `env:"SUPERVISOR_DONE_QUEUE_CLEANUP_INTERVAL,default=60m"`
}

type HTTPClientConfig struct {
	Timeout time.Duration `env:"HTTP_CLIENT_TIMEOUT,default=5s"`
}

type SinkConfig struct {
	DefaultDelay         time.Duration `env:"SINK_DEFAULT_DELAY,default=0"`
	DefaultMaxAttempts   int           `env:"SINK_DEFAULT_MAX_ATTEMPTS,default=3"`
	DefaultRetryInterval time.Duration `env:"SINK_DEFAULT_RETRY_AFTER,default=0"`
	// default retry exponential mutiplier is 1 (constant backoff)
	DefaultRetryExpMultiplier float64 `env:"SINK_DEFAULT_RETRY_EXP_MULTIPLIER,default=1"`
}

func InitAppConfig(ctx context.Context) (*AppConfig, error) {
	appConf := &AppConfig{}
	err := envconfig.Process(ctx, appConf)
	if err != nil {
		return nil, err
	}

	return appConf, nil
}
