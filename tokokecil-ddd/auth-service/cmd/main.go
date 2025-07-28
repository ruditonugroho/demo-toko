package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"auth-service/internal/auth/app"
	grpc2 "auth-service/internal/auth/delivery/grpc"
	"auth-service/internal/auth/delivery/grpc/pb"
	"auth-service/internal/auth/delivery/http"
	"auth-service/internal/auth/infra"
	"auth-service/pkg/hasher"
	"auth-service/pkg/jwt"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 2. Init MongoDB connection
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGO_URI & MONGO_DB must be set in .env")
	}

	maxPool, _ := strconv.ParseUint(getEnvOrDefault("MONGO_POOL_MAX", "50"), 10, 64)
	minPool, _ := strconv.ParseUint(getEnvOrDefault("MONGO_POOL_MIN", "5"), 10, 64)
	clientOpts := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(maxPool).
		SetMinPoolSize(minPool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("Gagal connect Mongo:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Println("Error disconnecting Mongo:", err)
		}
	}()
	log.Printf("✅ MongoDB Pooling aktif: min=%d max=%d\n", minPool, maxPool)
	userColl := client.Database(dbName).Collection("users")

	// 3. Init Infra Layer
	userRepo := &infra.MongoUserRepository{Collection: userColl}

	// 4. Init Shared Logic
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set in .env")
	}
	jwtManager := jwt.NewManager(jwtSecret)
	passwordHasher := hasher.NewBcrypt()

	// 5. Init App Layer
	authApp := app.NewAuthApp(userRepo, passwordHasher, jwtManager)

	// 6. Init HTTP Handler (Echo)
	authHandler := http.NewAuthHandler(authApp)
	go func() {
		e := echo.New()
		e.GET("/swagger/*", echoSwagger.WrapHandler)
		e.POST("/register", authHandler.Register)
		e.POST("/login", authHandler.Login)

		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		fmt.Println("HTTP Auth Service running at http://localhost:" + port)
		fmt.Println("Swagger UI: http://localhost:" + port + "/swagger/index.html")
		if err := e.Start(":" + port); err != nil {
			log.Fatal(err)
		}
	}()

	// 7. Init gRPC Handler & Interceptor
	grpcHandler := grpc2.NewAuthGRPCHandler(authApp)
	authInterceptor := grpc2.NewAuthInterceptor(jwtManager)

	// 8. Setup gRPC Server (ambil port dari env)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen gRPC: %v", err)
		}
		s := grpc.NewServer(
			grpc.UnaryInterceptor(authInterceptor.Unary()),
		)
		pb.RegisterAuthServiceServer(s, grpcHandler)
		log.Println("gRPC Auth Service running at :" + grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Block main goroutine (biar nggak langsung exit)
	select {}
}

// Helper ambil ENV dengan fallback (biar clean)
func getEnvOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
