// Command worker consumes async jobs from RabbitMQ (notifications, file
// processing, audit dispatch, retries).
//
// It exposes /healthz and /readyz on an HTTP port for orchestration, and runs
// its consumers in the background. Async jobs must carry the actor identity
// context of whoever created the job.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/audit"
	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/mq"
	effectapi "github.com/ting-boundless/boundless/services/worker/internal/effects"
	"github.com/ting-boundless/boundless/services/worker/internal/jobs"
	"github.com/ting-boundless/boundless/services/worker/internal/outbox"
	"github.com/ting-boundless/boundless/services/worker/internal/store"
)

const serviceName = "worker"
const migrationService = "worker-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	cfg := db.ConfigFromEnv("")
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
		if err := db.RunMigrations(cfg, migrationService); err != nil {
			log.Error("migrations failed", slog.Any("error", err))
			return
		}
	}

	rabbit := mq.Connect(ctx, log)
	if rabbit.Conn != nil {
		defer rabbit.Conn.Close()
	}

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}

	auditEmitter := audit.NewHTTPEmitter(audit.HTTPEmitterConfig{
		BaseURL: httpx.Env("AUDIT_SERVICE_URL", "http://127.0.0.1:8085"),
		Token:   internalToken,
	})

	if pg.DB != nil {
		if d := outbox.NewDispatcher(outbox.Config{
			Pool:   pg.DB.Pool(),
			Audit:  auditEmitter,
			Log:    log,
			Period: outboxPollInterval(),
		}); d != nil {
			go d.Run(ctx)
			log.Info("outbox dispatcher enabled")
		} else {
			log.Warn("outbox dispatcher disabled (postgres or audit-service unavailable)")
		}
	}

	health := httpx.NewHealth()
	mq.RegisterHealth(health, rabbit.Probe)
	db.RegisterHealth(health, "postgres", pg.Probe)

	var jobEffects *store.Effects
	if pg.DB != nil {
		jobEffects = store.NewEffects(pg.DB.Pool())
	}

	go startConsumers(ctx, log, rabbit.Conn, jobEffects)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("GET /internal/job-effects",
		httpx.InternalAuth(internalToken)(effectapi.New(jobEffects)),
	)

	h := httpx.Chain(mux, httpx.Recover(log), httpx.TraceContext)

	addr := httpx.Env("HTTP_ADDR", ":8086")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func outboxPollInterval() time.Duration {
	if s := httpx.Env("OUTBOX_POLL_INTERVAL", ""); s != "" {
		if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return 2 * time.Second
}

func startConsumers(ctx context.Context, log *slog.Logger, conn *amqp.Connection, effects *store.Effects) {
	if conn == nil {
		log.Info("rabbitmq consumers idle (not connected)")
		return
	}
	consumer := mq.NewConsumer(mq.ConsumerConfig{
		Conn:    conn,
		Queue:   mq.DefaultWorkQueue(),
		Handler: jobs.NewRouter(log, effects),
		Log:     log,
	})
	if consumer == nil {
		log.Warn("rabbitmq consumer not configured")
		return
	}
	go func() {
		if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
			log.Error("rabbitmq consumer stopped", slog.Any("error", err))
		}
	}()
}
