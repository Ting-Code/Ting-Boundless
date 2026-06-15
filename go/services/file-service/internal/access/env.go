package access

import (
	"strconv"
	"time"

	"github.com/ting-boundless/boundless/pkg/httpx"
)

// PresignSecondsFromEnv reads FILE_PRESIGN_SECONDS (default 3600).
func PresignSecondsFromEnv() time.Duration {
	if s := httpx.Env("FILE_PRESIGN_SECONDS", ""); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return time.Hour
}
