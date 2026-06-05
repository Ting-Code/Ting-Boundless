// Command worker consumes async jobs from RabbitMQ (notifications, file
// processing, audit dispatch, retries).
//
// It exposes /healthz and /readyz on an HTTP port for orchestration, and runs
// its consumers in the background. Async jobs must carry the actor identity
// context of whoever created the job.
package main

import (
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "worker"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()
	// TODO: health.Register(httpx.Check{Name: "rabbitmq", Probe: mq.Ping})

	// TODO: start RabbitMQ consumers here (with DLQ + idempotent handlers).
	go startConsumers(log)

	mux := http.NewServeMux()
	health.Handler(mux)

	h := httpx.Chain(mux, httpx.Recover(log))

	addr := httpx.Env("HTTP_ADDR", ":8081")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func startConsumers(log *slog.Logger) {
	log.Info("consumers starting (stub)")
	// TODO: connect to RABBITMQ_URL, declare queues + DLQ, process messages
	// idempotently, and propagate trace + identity context from the message.
}
