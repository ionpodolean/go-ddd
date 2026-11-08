package grpc

import (
	"context"
	"net"
	"testing"

	"go-ddd/internal/presentation/authguard"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func TestRateLimitUnaryInterceptor(t *testing.T) {
	limiter := authguard.New()
	ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1234}})
	interceptor := RateLimitUnaryInterceptor(limiter)
	handler := func(context.Context, any) (any, error) { return nil, nil }
	for i := 0; i < 21; i++ {
		_, err := interceptor(ctx, nil, &ggrpc.UnaryServerInfo{FullMethod: "/go_ddd.user.v1.UserService/Login"}, handler)
		if i < 20 && err != nil {
			t.Fatal(err)
		}
		if i == 20 && status.Code(err) != codes.ResourceExhausted {
			t.Fatalf("got %v", err)
		}
	}
	if _, err := interceptor(ctx, nil, &ggrpc.UnaryServerInfo{FullMethod: "/go_ddd.user.v1.UserService/GetProfile"}, handler); err != nil {
		t.Fatal(err)
	}
}
