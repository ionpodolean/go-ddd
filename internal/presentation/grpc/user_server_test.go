package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	userv1 "go-ddd/api/gen/go/user/v1"
	appUser "go-ddd/internal/application/user"
	domainUser "go-ddd/internal/domain/user"
	infraSecurity "go-ddd/internal/infrastructure/security"
	presentGRPC "go-ddd/internal/presentation/grpc"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockUserRepository struct{ users map[string]*domainUser.User }

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*domainUser.User)}
}
func (m *mockUserRepository) Create(_ context.Context, user *domainUser.User) error {
	user.ID = int64(len(m.users) + 1)
	m.users[user.Email] = user
	return nil
}
func (m *mockUserRepository) GetByEmail(_ context.Context, email string) (*domainUser.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, domainUser.ErrUserNotFound
	}
	return user, nil
}
func (m *mockUserRepository) GetByID(_ context.Context, id int64) (*domainUser.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, domainUser.ErrUserNotFound
}
func (m *mockUserRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := m.users[email]
	return ok, nil
}

func newTestClient(t *testing.T) (userv1.UserServiceClient, *infraSecurity.JWTService, *grpc.ClientConn) {
	t.Helper()
	jwtService := infraSecurity.NewJWTService("test-secret", time.Hour)
	userService := appUser.NewUserService(newMockUserRepository(), jwtService, nil)
	listener := bufconn.Listen(1024 * 1024)
	server := presentGRPC.NewServer(userService, jwtService)
	go func() { _ = server.Serve(listener) }()

	connection, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = connection.Close()
		server.GracefulStop()
		_ = listener.Close()
	})
	return userv1.NewUserServiceClient(connection), jwtService, connection
}

func TestUserServiceRPCs(t *testing.T) {
	client, _, _ := newTestClient(t)
	ctx := context.Background()

	registered, err := client.Register(ctx, &userv1.RegisterRequest{Email: "user@example.com", Password: "password123", FirstName: "Test", LastName: "User"})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.GetToken() == "" || registered.GetUser().GetCreatedAt() == nil {
		t.Fatal("Register() returned an incomplete authentication response")
	}

	loggedIn, err := client.Login(ctx, &userv1.LoginRequest{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	profile, err := client.GetProfile(metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+loggedIn.GetToken()), &userv1.GetProfileRequest{})
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}
	if profile.GetEmail() != "user@example.com" {
		t.Fatalf("GetProfile() email = %q, want user@example.com", profile.GetEmail())
	}
}

func TestUserServiceErrorMapping(t *testing.T) {
	client, jwtService, _ := newTestClient(t)
	ctx := context.Background()

	_, err := client.Register(ctx, &userv1.RegisterRequest{Email: "bad-email", Password: "password123", FirstName: "Test", LastName: "User"})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Errorf("invalid Register() code = %s, want %s", got, codes.InvalidArgument)
	}

	request := &userv1.RegisterRequest{Email: "user@example.com", Password: "password123", FirstName: "Test", LastName: "User"}
	if _, err := client.Register(ctx, request); err != nil {
		t.Fatalf("Register() setup error = %v", err)
	}
	if _, err := client.Register(ctx, request); status.Code(err) != codes.AlreadyExists {
		t.Errorf("duplicate Register() code = %s, want %s", status.Code(err), codes.AlreadyExists)
	}
	if _, err := client.Login(ctx, &userv1.LoginRequest{Email: request.GetEmail(), Password: "wrong-password"}); status.Code(err) != codes.Unauthenticated {
		t.Errorf("invalid Login() code = %s, want %s", status.Code(err), codes.Unauthenticated)
	}

	for _, authorization := range []string{"", "Basic token", "Bearer invalid-token", expiredToken(t)} {
		profileCtx := ctx
		if authorization != "" {
			profileCtx = metadata.AppendToOutgoingContext(ctx, "authorization", authorization)
		}
		if _, err := client.GetProfile(profileCtx, &userv1.GetProfileRequest{}); status.Code(err) != codes.Unauthenticated {
			t.Errorf("GetProfile(%q) code = %s, want %s", authorization, status.Code(err), codes.Unauthenticated)
		}
	}

	notFoundToken, err := jwtService.GenerateToken(999, "missing@example.com")
	if err != nil {
		t.Fatal(err)
	}
	missingCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+notFoundToken)
	if _, err := client.GetProfile(missingCtx, &userv1.GetProfileRequest{}); status.Code(err) != codes.NotFound {
		t.Errorf("missing GetProfile() code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestHealthAndReflection(t *testing.T) {
	_, _, conn := newTestClient(t)
	healthResponse, err := healthpb.NewHealthClient(conn).Check(context.Background(), &healthpb.HealthCheckRequest{})
	if err != nil || healthResponse.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("health check = %v, %v", healthResponse, err)
	}

	stream, err := reflectionpb.NewServerReflectionClient(conn).ServerReflectionInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Send(&reflectionpb.ServerReflectionRequest{MessageRequest: &reflectionpb.ServerReflectionRequest_ListServices{ListServices: ""}}); err != nil {
		t.Fatal(err)
	}
	response, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, service := range response.GetListServicesResponse().GetService() {
		if service.GetName() == "go_ddd.user.v1.UserService" {
			found = true
		}
	}
	if !found {
		t.Fatal("reflection response did not include UserService")
	}
}

func expiredToken(t *testing.T) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &infraSecurity.JWTClaims{
		UserID: 1,
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + signed
}
