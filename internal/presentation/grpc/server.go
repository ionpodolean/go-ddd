package grpc

import (
	userv1 "go-ddd/api/gen/go/user/v1"
	appUser "go-ddd/internal/application/user"
	infraSecurity "go-ddd/internal/infrastructure/security"
	"go-ddd/pkg/telemetry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(userService *appUser.UserService, jwtService *infraSecurity.JWTService) *grpc.Server {
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		RecoveryUnaryInterceptor,
		telemetry.GRPCObservabilityInterceptor,
		AuthUnaryInterceptor(jwtService),
	))

	userv1.RegisterUserServiceServer(server, NewUserServer(userService))
	healthpb.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server)
	return server
}
