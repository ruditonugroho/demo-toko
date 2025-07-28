package http

import (
	"auth-service/internal/auth/app"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler struct, DI via AuthApp (app layer)
type AuthHandler struct {
	App *app.AuthApp
}

// Constructor
func NewAuthHandler(app *app.AuthApp) *AuthHandler {
	return &AuthHandler{App: app}
}

// Register godoc
// @Summary Register user baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "data user baru"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	// (opsional) Validasi pakai validator
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Semua field wajib diisi"})
	}
	user, err := h.App.Register(c.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := AuthResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
	return c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body LoginRequest true "login"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	token, err := h.App.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	resp := AuthResponse{
		Email: req.Email,
		Token: token,
	}
	return c.JSON(http.StatusOK, resp)
}
