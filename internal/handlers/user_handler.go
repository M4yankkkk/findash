package handlers

import (
	"strconv"
	"strings"

	"github.com/M4yankkkk/findash/internal/middleware"
	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user management HTTP requests.
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers godoc
// @Summary      List all users
// @Description  Returns a paginated list of all users. Admin only.
// @Tags         users
// @Produce      json
// @Param        page      query     int  false  "Page number (default 1)"
// @Param        page_size query     int  false  "Page size (default 20, max 100)"
// @Success      200       {object}  utils.APIResponse
// @Failure      403       {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.ListUsers(page, pageSize)
	if err != nil {
		utils.InternalError(c, "failed to list users")
		return
	}

	utils.Paginated(c, users, total, page, pageSize)
}

// GetUser godoc
// @Summary      Get a user by ID
// @Description  Returns a single user. Admin only.
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUser(userID)
	if err != nil {
		utils.InternalError(c, "failed to get user")
		return
	}
	if user == nil {
		utils.NotFound(c, "user not found")
		return
	}

	utils.OK(c, "user retrieved", user)
}

// UpdateRole godoc
// @Summary      Update a user's role
// @Description  Changes the role of a user. Admin only. Cannot change own role.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "User ID"
// @Param        body  body      services.UpdateRoleInput  true  "New role"
// @Success      200   {object}  utils.APIResponse
// @Failure      400   {object}  utils.APIResponse
// @Failure      403   {object}  utils.APIResponse
// @Failure      404   {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /users/{id}/role [patch]
func (h *UserHandler) UpdateRole(c *gin.Context) {
	targetID := c.Param("id")
	requestingUserID := c.GetString(middleware.ContextKeyUserID)

	var input services.UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.UpdateRole(targetID, requestingUserID, models.Role(input.Role)); err != nil {
		switch err.Error() {
		case "user not found":
			utils.NotFound(c, err.Error())
		case "you cannot change your own role":
			utils.BadRequest(c, err.Error())
		default:
			utils.InternalError(c, "failed to update role")
		}
		return
	}

	utils.OK(c, "role updated successfully", nil)
}

// ListViewers returns paginated users with viewer role for assignment flows.
func (h *UserHandler) ListViewers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	viewers, total, err := h.userService.ListViewers(page, pageSize)
	if err != nil {
		utils.InternalError(c, "failed to list viewers")
		return
	}

	utils.Paginated(c, viewers, total, page, pageSize)
}

// GetViewerVisibility returns entry IDs visible to a specific viewer.
func (h *UserHandler) GetViewerVisibility(c *gin.Context) {
	viewerID := c.Param("id")

	entryIDs, err := h.userService.GetViewerVisibility(viewerID)
	if err != nil {
		switch err.Error() {
		case "viewer not found":
			utils.NotFound(c, err.Error())
		case "target user is not a viewer":
			utils.BadRequest(c, err.Error())
		default:
			utils.InternalError(c, "failed to fetch viewer visibility")
		}
		return
	}

	utils.OK(c, "viewer visibility retrieved", gin.H{"entry_ids": entryIDs})
}

// UpdateViewerVisibility replaces all visible entries for a viewer.
func (h *UserHandler) UpdateViewerVisibility(c *gin.Context) {
	viewerID := c.Param("id")
	actorID := c.GetString(middleware.ContextKeyUserID)
	actorRole := models.Role(c.GetString(middleware.ContextKeyRole))

	var input services.UpdateViewerVisibilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.ReplaceViewerVisibility(viewerID, input.EntryIDs, actorID, actorRole); err != nil {
		switch err.Error() {
		case "viewer not found":
			utils.NotFound(c, err.Error())
		case "target user is not a viewer":
			utils.BadRequest(c, err.Error())
		case "forbidden entry assignment":
			utils.Forbidden(c, err.Error())
		default:
			if strings.Contains(err.Error(), "entry not found:") {
				utils.BadRequest(c, err.Error())
				return
			}
			utils.InternalError(c, "failed to update viewer visibility")
		}
		return
	}

	utils.OK(c, "viewer visibility updated", nil)
}
