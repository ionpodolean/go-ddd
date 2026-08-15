package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ObservabilityMiddleware wraps HTTP handlers with OpenTelemetry Tracing, Prometheus Metrics, and Zerolog logging.
func ObservabilityMiddleware(next http.Handler) http.Handler {
	// Wrap handler with otelhttp for automatic distributed tracing span creation
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := r.Context()

		// Track active requests gauge
		if HTTPActiveRequests != nil {
			HTTPActiveRequests.Add(ctx, 1)
			defer HTTPActiveRequests.Add(ctx, -1)
		}

		wrapper := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute downstream handlers
		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		durationSec := duration.Seconds()

		// Context attributes
		statusStr := fmt.Sprintf("%d", wrapper.statusCode)
		attrOpt := metric.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
			attribute.String("http.status_code", statusStr),
		)

		// Record metrics
		if HTTPRequestsTotal != nil {
			HTTPRequestsTotal.Add(ctx, 1, attrOpt)
		}
		if HTTPRequestDurationSeconds != nil {
			HTTPRequestDurationSeconds.Record(ctx, durationSec, attrOpt)
		}

		// Log request with Zerolog + Trace context correlation
		logger := Logger(ctx)
		var logEvent *zerolog.Event
		if wrapper.statusCode >= 500 {
			logEvent = logger.Error()
		} else if wrapper.statusCode >= 400 {
			logEvent = logger.Warn()
		} else {
			logEvent = logger.Info()
		}

		logEvent.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapper.statusCode).
			Float64("duration_ms", float64(duration.Microseconds())/1000.0).
			Str("user_agent", r.UserAgent()).
			Str("remote_ip", r.RemoteAddr).
			Msg("HTTP request processed")

		// Annotate current OTel span with status and attributes
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.target", r.URL.Path),
				attribute.Int("http.status_code", wrapper.statusCode),
			)
		}
	}), "HTTP Server")

	return otelHandler
}
