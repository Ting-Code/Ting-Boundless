// Command file-service handles uploads/downloads over S3-compatible storage
// (MinIO / Aliyun OSS). Identity comes from the Gateway via identity.Middleware.
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/errs"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/identity"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/storage"
)

const serviceName = "file-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
	}
	s3 := storage.Connect(ctx, log)

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)
	storage.RegisterHealth(health, s3.Probe)

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("POST /v1/files/", identity.Middleware(http.HandlerFunc(handleUpload)))

	internalToken := httpx.Env("INTERNAL_API_TOKEN", "")

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
	)

	addr := httpx.Env("HTTP_ADDR", ":8083")
	if err := httpx.New(addr, h, log).Run(); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	id, _ := identity.FromContext(r.Context())
	errs.Write(w, id.RequestID, errs.New(http.StatusNotImplemented, "not_implemented", "file upload not implemented"))
}
