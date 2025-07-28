package grpc

import (
	"context"
	"strings"

	"auth-service/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor adalah unary interceptor untuk validasi JWT di setiap RPC
type AuthInterceptor struct {
	JWTManager *jwt.Manager
	// Bisa tambahkan allowListMethod jika mau skip auth di method tertentu
}

// Constructor
func NewAuthInterceptor(jwtManager *jwt.Manager) *AuthInterceptor {
	return &AuthInterceptor{
		JWTManager: jwtManager,
	}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Bypass auth untuk endpoint tertentu (login/register)
		if strings.HasSuffix(info.FullMethod, "Login") || strings.HasSuffix(info.FullMethod, "Register") {
			return handler(ctx, req)
		}

		// Extract JWT dari metadata (key: authorization)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token not provided")
		}

		// Format token: "Bearer <token>"
		tokenStr := values[0]
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		_, err := a.JWTManager.VerifyToken(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
		}
		// lanjutkan ke handler utama jika token valid
		return handler(ctx, req)
	}
}
