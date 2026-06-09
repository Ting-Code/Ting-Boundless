package config

import (
	"os"

	"github.com/joho/godotenv"
)

// LoadEnvFile loads .env from the working directory or repo root (../.env when
// running from go/). Production injects env at deploy time.
func LoadEnvFile() {
	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
	}
}
