package lib

import (
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func InitRedisClient(conf *AppConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(conf.Redis.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse redis url")
	}

	client := redis.NewClient(opt)

	return client, nil
}
