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

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/mq"
)

const serviceName = "worker"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	rabbit := mq.Connect(ctx, log)
	if rabbit.Conn != nil {
		defer rabbit.Conn.Close()
	}

	health := httpx.NewHealth()
	mq.RegisterHealth(health, rabbit.Probe)

	go startConsumers(log, rabbit.Conn)

	mux := http.NewServeMux()
	health.Handler(mux)

	h := httpx.Chain(mux, httpx.Recover(log))

	addr := httpx.Env("HTTP_ADDR", ":8086")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func startConsumers(log *slog.Logger, conn *amqp.Connection) {
	if conn == nil {
		log.Info("consumers idle (rabbitmq not connected)")
		return
	}
	log.Info("consumers starting (stub)")
}
