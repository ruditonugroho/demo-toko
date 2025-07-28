package config

import (
	"log"
	"os"

	authpb "gateway-service/pb"
	productpb "gateway-service/pb"

	"google.golang.org/grpc"
)

type GRPCClients struct {
	AuthClient    authpb.AuthServiceClient
	ProductClient productpb.ProductServiceClient
}

func NewGRPCClients() *GRPCClients {
	authConn, err := grpc.Dial(os.Getenv("AUTH_SERVICE_URL"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	productConn, err := grpc.Dial(os.Getenv("PRODUCT_SERVICE_URL"), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	return &GRPCClients{
		AuthClient:    authpb.NewAuthServiceClient(authConn),
		ProductClient: productpb.NewProductServiceClient(productConn),
	}
}
