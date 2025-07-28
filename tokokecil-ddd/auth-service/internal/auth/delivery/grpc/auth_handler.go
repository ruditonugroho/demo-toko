package grpc

import (
	"context"

	"auth-service/internal/auth/app"
	"auth-service/internal/auth/delivery/grpc/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler gRPC, implementasi pb.AuthServiceServer
type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	App *app.AuthApp
}

// Constructor
func NewAuthGRPCHandler(app *app.AuthApp) *AuthGRPCHandler {
	return &AuthGRPCHandler{App: app}
}

// Register user baru via gRPC
func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	user, err := h.App.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, errToGRPC(err)
	}
	// Kamu bisa generate token langsung jika ingin login setelah register
	token, _ := h.App.JWTManager.GenerateToken(user.ID, user.Email)
	return &pb.AuthResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Token: token,
	}, nil
}

// Login user via gRPC
func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	token, err := h.App.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, errToGRPC(err)
	}
	return &pb.AuthResponse{
		Email: req.GetEmail(),
		Token: token,
	}, nil
}

// Helper: konversi error aplikasi ke status gRPC
func errToGRPC(err error) error {
	switch err {
	case app.ErrValidation:
		return status.Error(codes.InvalidArgument, err.Error())
	case app.ErrEmailExist:
		return status.Error(codes.AlreadyExists, err.Error())
	case app.ErrUnauthorized:
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
