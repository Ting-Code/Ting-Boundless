// Command file-service handles uploads/downloads over S3-compatible storage
// (MinIO / Aliyun OSS). Identity comes from the Gateway via identity.Middleware.
package main

import (
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
)

const serviceName = "file-service"

func main() {
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	health := httpx.NewHealth()
	// TODO: health.Register(httpx.Check{Name: "s3", Probe: store.Ping})

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("POST /v1/files/", identity.Middleware(http.HandlerFunc(handleUpload)))

	h := httpx.Chain(mux,
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8080")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleUpload(w http.ResponseWriter, _ *http.Request) {
	// TODO: stream to S3-compatible storage, record metadata in app_db.
	httpx.JSON(w, http.StatusNotImplemented, map[string]string{"code": "not_implemented"})
}
