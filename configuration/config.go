package configuration

import (
	"github.com/joho/godotenv"
	"os"
)

type Config interface {
	Get(key string) string
}

type configImpl struct {
}

func (config *configImpl) Get(key string) string {
	return os.Getenv(key)
}

func New(filenames ...string) Config {
	_ = godotenv.Load(filenames...)
	return &configImpl{}
}
