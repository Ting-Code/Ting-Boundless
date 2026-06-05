package config

import "github.com/joho/godotenv"

// LoadEnvFile loads .env from the working directory when present.
// Production injects env at deploy time; this is a local-dev convenience only.
func LoadEnvFile() {
	_ = godotenv.Load(".env")
}
