package config

import (
	"errors"
	"io/fs"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads variables from the .env file in the working directory.
// A missing file is not an error: in Docker the variables come from the environment.
// Variables that are already set in the environment are not overridden.
func LoadDotEnv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// Get returns the value of the environment variable key,
// or def if the variable is not set or empty.
func Get(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
