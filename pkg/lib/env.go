package lib

import (
	"os"

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
	return LoadEnvFromFile(".env", true)
}

func LoadEnvFromFile(filename string, skipIfNotExists bool) error {
	err := godotenv.Load(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && skipIfNotExists {
			// file does not exist
			return nil
		}

		return errors.Wrapf(err, "error processing dotenv config. file: %s", filename)
	}

	return nil
}
