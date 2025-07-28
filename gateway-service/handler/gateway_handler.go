package handler

import (
	"context"
	"net/http"

	"gateway-service/config"
	authpb "gateway-service/pb"
	productpb "gateway-service/pb"

	"github.com/labstack/echo/v4"
)

type GatewayHandler struct {
	GRPC *config.GRPCClients
}

func NewGatewayHandler(grpcClients *config.GRPCClients) *GatewayHandler {
	return &GatewayHandler{GRPC: grpcClients}
}

func (h *GatewayHandler) Register(c echo.Context) error {
	var req authpb.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	resp, err := h.GRPC.AuthClient.Register(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *GatewayHandler) Login(c echo.Context) error {
	var req authpb.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	resp, err := h.GRPC.AuthClient.Login(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *GatewayHandler) GetAllProducts(c echo.Context) error {
	resp, err := h.GRPC.ProductClient.GetAllProducts(context.Background(), &productpb.Empty{})
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, resp.Products)
}

func (h *GatewayHandler) CreateProduct(c echo.Context) error {
	var req productpb.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	resp, err := h.GRPC.ProductClient.CreateProduct(context.Background(), &req)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, resp.Product)
}
