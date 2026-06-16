// Command file-service handles uploads/downloads over S3-compatible storage
// (MinIO / Aliyun OSS). Identity comes from the Gateway via identity.Middleware.
package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/db"
	"github.com/ting-boundless/boundless/pkg/httpx"
	"github.com/ting-boundless/boundless/pkg/logger"
	"github.com/ting-boundless/boundless/pkg/storage"
	"github.com/ting-boundless/boundless/services/file-service/internal/access"
	"github.com/ting-boundless/boundless/services/file-service/internal/store"
	"github.com/ting-boundless/boundless/services/file-service/internal/upload"
)

const serviceName = "file-service"

func main() {
	config.LoadEnvFile()
	log := logger.New(serviceName, httpx.Env("LOG_LEVEL", "info"))
	slog.SetDefault(log)

	ctx := context.Background()
	cfg := db.ConfigFromEnv("")
	pg := db.Connect(ctx, log, "")
	if pg.DB != nil {
		defer pg.DB.Close()
		if err := db.RunMigrations(cfg, serviceName); err != nil {
			log.Error("migrations failed", slog.Any("error", err))
			return
		}
	}

	s3Probe := storage.Connect(ctx, log)
	s3Client, err := storage.NewClient(ctx, storage.ConfigFromEnv())
	if err != nil {
		log.Error("s3 client init failed", slog.Any("error", err))
		return
	}
	if s3Client == nil {
		log.Warn("s3 uploads disabled (set S3_ENDPOINT, keys, and bucket)")
	} else {
		log.Info("s3 client ready", slog.String("bucket", s3Client.Bucket()))
	}

	var files *store.Files
	if pg.DB != nil {
		files = store.NewFiles(pg.DB.Pool())
	}

	health := httpx.NewHealth()
	db.RegisterHealth(health, "postgres", pg.Probe)
	storage.RegisterHealth(health, s3Probe.Probe)

	uploadHandler := upload.New(upload.Config{
		Files:    files,
		S3:       s3Client,
		MaxBytes: upload.MaxBytesFromEnv(),
		Log:      log,
	})
	accessCfg := access.Config{
		Files:         files,
		S3:            s3Client,
		DefaultExpiry: access.PresignSecondsFromEnv(),
		Log:           log,
	}

	mux := http.NewServeMux()
	health.Handler(mux)
	mux.Handle("POST /v1/files/", httpx.TrustedAuth(uploadHandler))
	mux.Handle("GET /v1/files/", httpx.TrustedAuth(access.NewList(accessCfg)))
	mux.Handle("GET /v1/files/{id}", httpx.TrustedAuth(access.NewMeta(accessCfg)))
	mux.Handle("GET /v1/files/{id}/download", httpx.TrustedAuth(access.NewDownload(accessCfg)))
	mux.Handle("GET /v1/files/{id}/url", httpx.TrustedAuth(access.NewURL(accessCfg)))
	mux.Handle("DELETE /v1/files/{id}", httpx.TrustedAuth(access.NewDelete(accessCfg)))

	internalToken, ok := httpx.LoadInternalToken(log)
	if !ok {
		return
	}

	h := httpx.Chain(mux,
		httpx.GatewayTrust(internalToken),
		httpx.RequestID,
		httpx.Recover(log),
		httpx.AccessLog(log),
		httpx.TraceContext,
	)

	addr := httpx.Env("HTTP_ADDR", ":8083")
	if err := httpx.RunService(addr, serviceName, h, log); err != nil {
		log.Error("server error", slog.Any("error", err))
	}
}
