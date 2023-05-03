package lib

import (
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

type AppEnv string

const (
	AppEnvTest        AppEnv = "test"
	AppEnvDevelopment AppEnv = "development"
	AppEnvProduction  AppEnv = "production"
)

func LoadEnv() error {
	return LoadEnvFromFile(".env")
}

func LoadEnvFromFile(filename string) error {
	err := godotenv.Load(filename)
	if err != nil {
		return errors.Wrapf(err, "error loading %s file", filename)
	}

	return nil
}
