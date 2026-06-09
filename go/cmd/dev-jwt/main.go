// Command dev-jwt prints a local HS256 JWT for Gateway testing (GATEWAY_DEV_JWT_SECRET).
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ting-boundless/boundless/pkg/auth"
	"github.com/ting-boundless/boundless/pkg/config"
)

func main() {
	config.LoadEnvFile()
	cfg := auth.ConfigFromEnv()
	userID := "dev-user"
	if len(os.Args) > 1 {
		userID = os.Args[1]
	}
	tok, err := auth.DevToken(cfg, userID, "dev-tenant", []string{"user"}, 24*time.Hour)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(tok)
}
