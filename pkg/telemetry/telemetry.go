package telemetry

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

var (
	tracer trace.Tracer
	meter  metric.Meter

	// Global Metrics
	HTTPRequestsTotal          metric.Int64Counter
	HTTPRequestDurationSeconds metric.Float64Histogram
	HTTPActiveRequests         metric.Int64UpDownCounter
	GRPCRequestsTotal          metric.Int64Counter
	GRPCRequestDurationSeconds metric.Float64Histogram
	GRPCActiveRequests         metric.Int64UpDownCounter

	once sync.Once
)

// InitTelemetry initializes Zerolog logger and OpenTelemetry Providers (Traces and Metrics).
func InitTelemetry(ctx context.Context) (func(context.Context) error, error) {
	initZerolog()

	serviceName := getEnv("OTEL_SERVICE_NAME", "go-ddd-service")
	serviceVersion := getEnv("OTEL_SERVICE_VERSION", "1.0.0")
	env := getEnv("ENVIRONMENT", "development")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otel resource: %w", err)
	}

	// Propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var shutdownFuncs []func(context.Context) error

	// Trace Provider
	var tpOptions []sdktrace.TracerProviderOption
	tpOptions = append(tpOptions, sdktrace.WithResource(res))

	if otlpEndpoint != "" {
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(otlpEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create OTLP trace exporter, running without remote tracing")
		} else {
			tpOptions = append(tpOptions, sdktrace.WithBatcher(traceExporter))
		}
	}

	tp := sdktrace.NewTracerProvider(tpOptions...)
	otel.SetTracerProvider(tp)
	tracer = tp.Tracer(serviceName)
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)

	// Meter Provider with Prometheus exporter & OTLP Metric Exporter
	promExporter, err := promexporter.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	mpOptions := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
	}

	if otlpEndpoint != "" {
		metricExporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(otlpEndpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create OTLP metric exporter")
		} else {
			mpOptions = append(mpOptions, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)))
		}
	}

	mp := sdkmetric.NewMeterProvider(mpOptions...)
	otel.SetMeterProvider(mp)
	meter = mp.Meter(serviceName)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)

	// Initialize metrics
	if err := initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	log.Info().
		Str("service", serviceName).
		Str("otlp_endpoint", otlpEndpoint).
		Msg("Telemetry initialized successfully")

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			if shutErr := fn(ctx); shutErr != nil {
				err = shutErr
			}
		}
		return err
	}

	return shutdown, nil
}

func initZerolog() {
	zerolog.TimeFieldFormat = time.RFC3339
	logLevelStr := getEnv("LOG_LEVEL", "info")

	level, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if getEnv("LOG_PRETTY", "false") == "true" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func initMetrics() error {
	var err error
	HTTPRequestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests processed"),
	)
	if err != nil {
		return err
	}

	HTTPRequestDurationSeconds, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	HTTPActiveRequests, err = meter.Int64UpDownCounter(
		"http_active_requests",
		metric.WithDescription("Number of active HTTP requests currently being processed"),
	)
	if err != nil {
		return err
	}

	GRPCRequestsTotal, err = meter.Int64Counter(
		"grpc_server_requests_total",
		metric.WithDescription("Total number of gRPC requests processed"),
	)
	if err != nil {
		return err
	}

	GRPCRequestDurationSeconds, err = meter.Float64Histogram(
		"grpc_server_request_duration_seconds",
		metric.WithDescription("gRPC request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	GRPCActiveRequests, err = meter.Int64UpDownCounter(
		"grpc_server_active_requests",
		metric.WithDescription("Number of active gRPC requests currently being processed"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Tracer returns the global OpenTelemetry tracer.
func Tracer() trace.Tracer {
	if tracer == nil {
		return otel.Tracer("go-ddd-service")
	}
	return tracer
}

// Logger returns a Zerolog logger enriched with trace_id and span_id from context if available.
func Logger(ctx context.Context) zerolog.Logger {
	logger := log.Logger
	if ctx == nil {
		return logger
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return logger.With().
			Str("trace_id", spanCtx.TraceID().String()).
			Str("span_id", spanCtx.SpanID().String()).
			Logger()
	}

	return logger
}

// PrometheusHandler returns an http.Handler for Prometheus metrics scraping.
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
