package handlers

import (
	"strconv"

	"github.com/M4yankkkk/findash/internal/middleware"
	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsService *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// GetSummary godoc
// @Summary      Get financial summary
// @Description  Returns total income, total expenses, and net balance.
//
//	Admins see global data; others see only their own.
//
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  utils.APIResponse
// @Failure      403  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /analytics/summary [get]
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	summary, err := h.analyticsService.GetSummary(userID, role)
	if err != nil {
		utils.InternalError(c, "failed to fetch summary")
		return
	}

	utils.OK(c, "summary retrieved", summary)
}

// GetByCategory godoc
// @Summary      Get breakdown by category
// @Description  Returns income and expense totals grouped by category, ordered by total descending.
//
//	Admins see global data; others see only their own.
//
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  utils.APIResponse
// @Failure      403  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /analytics/by-category [get]
func (h *AnalyticsHandler) GetByCategory(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	breakdown, err := h.analyticsService.GetByCategory(userID, role)
	if err != nil {
		utils.InternalError(c, "failed to fetch category breakdown")
		return
	}

	utils.OK(c, "category breakdown retrieved", breakdown)
}

// GetTrend godoc
// @Summary      Get monthly income vs expense trend
// @Description  Returns month-by-month income and expense totals for the last N months (default 6, max 24).
//
//	Admins see global data; others see only their own.
//
// @Tags         analytics
// @Produce      json
// @Param        months  query     int  false  "Number of months to include (default 6, max 24)"
// @Success      200     {object}  utils.APIResponse
// @Failure      403     {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /analytics/trend [get]
func (h *AnalyticsHandler) GetTrend(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))

	trend, err := h.analyticsService.GetTrend(userID, role, months)
	if err != nil {
		utils.InternalError(c, "failed to fetch trend data")
		return
	}

	utils.OK(c, "trend data retrieved", trend)
}
