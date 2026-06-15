package otel
import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/logger"
)

// AttachLogExport adds an OTLP slog handler when logs export is enabled.
// JSON stdout logging is preserved via fan-out. Returns the logger unchanged when disabled.
func AttachLogExport(ctx context.Context, serviceName string, base *slog.Logger) (*slog.Logger, func(context.Context) error, error) {
	if !logsExportEnabled() {
		return base, func(context.Context) error { return nil }, nil
	}

	endpoint := strings.TrimSpace(envOr("OTEL_EXPORTER_OTLP_ENDPOINT", ""))
	if endpoint == "" || config.IsPlaceholder(endpoint) {
		return base, func(context.Context) error { return nil }, nil
	}

	name := strings.TrimSpace(envOr("OTEL_SERVICE_NAME", ""))
	if name == "" || name == "unset" {
		name = serviceName
	}

	exporter, err := newLogExporter(ctx, endpoint)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("otel log resource: %w", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	otelHandler := otelslog.NewHandler(name, otelslog.WithLoggerProvider(provider))
	level := envOr("LOG_LEVEL", "info")
	stdout := logger.New(serviceName, level).Handler()
	combined := logger.NewFanout(stdout, otelHandler)
	out := logger.WithHandler(serviceName, level, combined)

	if base != nil {
		base.Info("otel logs export enabled",
			slog.String("endpoint", endpoint),
			slog.String("service", name),
		)
	}

	return out, provider.Shutdown, nil
}

func logsExportEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(envOr("OTEL_LOGS_EXPORTER", "otlp"))) {
	case "none", "false", "off":
		return false
	default:
		return true
	}
}

func newLogExporter(ctx context.Context, endpoint string) (sdklog.Exporter, error) {
	proto := strings.ToLower(strings.TrimSpace(envOr("OTEL_EXPORTER_OTLP_PROTOCOL", "")))
	if proto == "" {
		if strings.Contains(endpoint, ":4317") {
			proto = "grpc"
		} else {
			proto = "http/protobuf"
		}
	}

	switch proto {
	case "grpc":
		return otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(trimScheme(endpoint)),
			otlploggrpc.WithInsecure(),
		)
	case "http/protobuf", "http":
		return otlploghttp.New(ctx,
			otlploghttp.WithEndpoint(trimScheme(endpoint)),
			otlploghttp.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_PROTOCOL %q", proto)
	}
}
