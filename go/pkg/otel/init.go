// Package otel initializes OpenTelemetry tracing (OTLP export) for Go services.
package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/ting-boundless/boundless/pkg/config"
)

var errDisabled = errors.New("otel disabled")

// InitFromEnv configures the global TracerProvider when OTEL_EXPORTER_OTLP_ENDPOINT
// is set. Returns a shutdown function (no-op when disabled).
func InitFromEnv(ctx context.Context, serviceName string, log *slog.Logger) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(envOr("OTEL_EXPORTER_OTLP_ENDPOINT", ""))
	if endpoint == "" || config.IsPlaceholder(endpoint) {
		if log != nil {
			log.Info("otel tracing disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)")
		}
		return func(context.Context) error { return nil }, nil
	}

	name := strings.TrimSpace(envOr("OTEL_SERVICE_NAME", ""))
	if name == "" || name == "unset" {
		name = serviceName
	}

	exporter, err := newTraceExporter(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if log != nil {
		log.Info("otel tracing enabled",
			slog.String("endpoint", endpoint),
			slog.String("service", name),
		)
	}

	return tp.Shutdown, nil
}

func newTraceExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
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
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(trimScheme(endpoint)),
			otlptracegrpc.WithInsecure(),
		)
	case "http/protobuf", "http":
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(trimScheme(endpoint)),
			otlptracehttp.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_PROTOCOL %q", proto)
	}
}

func trimScheme(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return strings.TrimSuffix(endpoint, "/")
}

// WrapHandler instruments inbound HTTP when a TracerProvider is configured.
func WrapHandler(handler http.Handler, operation string) http.Handler {
	if handler == nil {
		return nil
	}
	if operation == "" {
		operation = "http"
	}
	return otelhttp.NewHandler(handler, operation,
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// NewHTTPClient returns an HTTP client that propagates trace context on outbound calls.
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
