package util

import (
	"github.com/spf13/cast"
	"os"
)

var (
	Environment string
	Production  = "production"
	Dev         = "dev"
	useDatadog  = cast.ToBool(GetEnv("USE_DATADOG", "false"))
)

func init() {
	Environment = GetEnv("EVO_ENV", Dev)
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func IsProduction() bool {
	return Environment == Production
}

func UseDatadog() bool {
	return useDatadog
}
