package main

import (
	"log"
	"os"

	"gateway-service/config"
	"gateway-service/handler"
	"gateway-service/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	config.LoadEnv()

	grpcClients := config.NewGRPCClients()
	h := handler.NewGatewayHandler(grpcClients)

	e := echo.New()

	// AUTH endpoints (proxy ke Auth Service via gRPC)
	e.POST("/register", h.Register)
	e.POST("/login", h.Login)

	// PRODUCT endpoints (proxy ke Product Service via gRPC, protected by JWT)
	productGroup := e.Group("/products", middleware.JWTAuth)
	productGroup.GET("", h.GetAllProducts)
	productGroup.POST("", h.CreateProduct)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Println("Gateway Service running at http://localhost:" + port)
	e.Logger.Fatal(e.Start(":" + port))
}
