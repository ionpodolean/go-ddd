package telemetry

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCObservabilityInterceptor records traces, metrics, and structured logs for unary gRPC requests.
func GRPCObservabilityInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	service, method := grpcMethodParts(info.FullMethod)
	attrs := []attribute.KeyValue{
		attribute.String("rpc.system", "grpc"),
		attribute.String("rpc.service", service),
		attribute.String("rpc.method", method),
	}

	ctx, span := Tracer().Start(ctx, info.FullMethod)
	defer span.End()
	start := time.Now()
	if GRPCActiveRequests != nil {
		GRPCActiveRequests.Add(ctx, 1, metric.WithAttributes(attrs...))
		defer GRPCActiveRequests.Add(ctx, -1, metric.WithAttributes(attrs...))
	}

	response, err := handler(ctx, req)
	grpcCode := status.Code(err)
	attrs = append(attrs, attribute.String("rpc.grpc.status_code", grpcCode.String()))
	duration := time.Since(start).Seconds()
	options := metric.WithAttributes(attrs...)
	if GRPCRequestsTotal != nil {
		GRPCRequestsTotal.Add(ctx, 1, options)
	}
	if GRPCRequestDurationSeconds != nil {
		GRPCRequestDurationSeconds.Record(ctx, duration, options)
	}

	if err != nil && grpcCode != grpcCodes.OK {
		span.RecordError(err)
		span.SetStatus(codes.Error, grpcCode.String())
	}
	logger := Logger(ctx).With().Str("rpc_service", service).Str("rpc_method", method).Str("grpc_code", grpcCode.String()).Float64("duration_ms", float64(time.Since(start).Microseconds())/1000).Logger()
	if grpcCode == grpcCodes.OK {
		logger.Info().Msg("gRPC request processed")
	} else if grpcCode == grpcCodes.Internal {
		logger.Error().Msg("gRPC request processed")
	} else {
		logger.Warn().Msg("gRPC request processed")
	}
	return response, err
}

func grpcMethodParts(fullMethod string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	if len(parts) != 2 {
		return "unknown", fullMethod
	}
	return parts[0], parts[1]
}
