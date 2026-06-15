package ratelimit

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// redisLimiter enforces per-IP fixed-window limits shared across Gateway replicas.
type redisLimiter struct {
	rdb    *goredis.Client
	auth   windowLimit
	general windowLimit
}

type windowLimit struct {
	max    int64
	period time.Duration
}

func newRedisLimiter(rdb *goredis.Client, cfg Config) *redisLimiter {
	return &redisLimiter{
		rdb: rdb,
		auth: windowLimit{
			max:    int64(limitPerWindow(cfg.AuthRPS, cfg.AuthBurst)),
			period: time.Second,
		},
		general: windowLimit{
			max:    int64(limitPerWindow(cfg.GeneralRPS, cfg.GeneralBurst)),
			period: time.Second,
		},
	}
}

func (l *redisLimiter) allow(ctx context.Context, zone, ip string) (bool, error) {
	if l == nil || l.rdb == nil {
		return true, nil
	}
	wl := l.general
	if zone == "auth" {
		wl = l.auth
	}
	if wl.max <= 0 {
		return true, nil
	}
	key := fmt.Sprintf("gw:rl:%s:%s:%d", zone, ip, time.Now().Unix()/int64(wl.period.Seconds()))
	n, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = l.rdb.Expire(ctx, key, wl.period+time.Second).Err()
	}
	return n <= wl.max, nil
}

func limitPerWindow(rps float64, burst int) int {
	if burst > 0 {
		return burst
	}
	if rps <= 0 {
		return 0
	}
	n := int(rps)
	if n < 1 {
		n = 1
	}
	return n
}
