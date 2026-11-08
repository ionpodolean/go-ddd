package grpc

import (
	"context"
	"strings"

	infraSecurity "go-ddd/internal/infrastructure/security"
	"go-ddd/internal/presentation/authguard"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type claimsContextKey struct{}

func RateLimitUnaryInterceptor(limiter *authguard.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == "/go_ddd.user.v1.UserService/Register" || info.FullMethod == "/go_ddd.user.v1.UserService/Login" {
			address := "unknown"
			if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
				address = p.Addr.String()
			}
			if !limiter.Allow(address) {
				return nil, status.Error(codes.ResourceExhausted, "too many authentication attempts")
			}
		}
		return handler(ctx, req)
	}
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(*infraSecurity.JWTClaims)
	if !ok || claims == nil {
		return 0, false
	}
	return claims.UserID, true
}

func AuthUnaryInterceptor(jwtService *infraSecurity.JWTService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod != "/go_ddd.user.v1.UserService/GetProfile" {
			return handler(ctx, req)
		}

		values := metadata.ValueFromIncomingContext(ctx, "authorization")
		if len(values) != 1 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
		}
		parts := strings.Fields(values[0])
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization metadata format")
		}

		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			log.Warn().Err(err).Str("rpc_method", info.FullMethod).Msg("Invalid or expired gRPC token")
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		return handler(context.WithValue(ctx, claimsContextKey{}, claims), req)
	}
}

func RecoveryUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Error().Interface("panic", recovered).Str("rpc_method", info.FullMethod).Msg("gRPC panic recovered")
			response = nil
			err = status.Error(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}
