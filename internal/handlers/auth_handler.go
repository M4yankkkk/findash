package handlers

import (
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account. The first registered user is automatically assigned the admin role.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      services.RegisterInput  true  "Registration details"
// @Success      201   {object}  utils.APIResponse
// @Failure      400   {object}  utils.APIResponse
// @Failure      409   {object}  utils.APIResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input services.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Register(input)
	if err != nil {
		// Email already registered is a conflict, not a bad request
		if err.Error() == "email already registered" {
			utils.Conflict(c, err.Error())
			return
		}
		utils.InternalError(c, "registration failed")
		return
	}

	utils.Created(c, "registration successful", resp)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      services.LoginInput  true  "Login credentials"
// @Success      200   {object}  utils.APIResponse
// @Failure      400   {object}  utils.APIResponse
// @Failure      401   {object}  utils.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authService.Login(input)
	if err != nil {
		if err.Error() == "account is inactive" {
			utils.Forbidden(c, "account is inactive")
			return
		}
		// Always return 401 for auth failures — never reveal whether email exists
		utils.Unauthorized(c, "invalid email or password")
		return
	}

	utils.OK(c, "login successful", resp)
}
