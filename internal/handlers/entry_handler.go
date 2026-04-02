package handlers

import (
	"strconv"
	"time"

	"github.com/M4yankkkk/findash/internal/middleware"
	"github.com/M4yankkkk/findash/internal/models"
	"github.com/M4yankkkk/findash/internal/services"
	"github.com/M4yankkkk/findash/pkg/utils"
	"github.com/gin-gonic/gin"
)

// EntryHandler handles financial entry HTTP requests.
type EntryHandler struct {
	entryService *services.EntryService
}

// NewEntryHandler creates a new EntryHandler.
func NewEntryHandler(entryService *services.EntryService) *EntryHandler {
	return &EntryHandler{entryService: entryService}
}

// CreateEntry godoc
// @Summary      Create a financial entry
// @Description  Creates a new income or expense entry. Manager and Admin only.
// @Tags         entries
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateEntryInput  true  "Entry details"
// @Success      201   {object}  utils.APIResponse
// @Failure      400   {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /entries [post]
func (h *EntryHandler) CreateEntry(c *gin.Context) {
	var input models.CreateEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	entry, err := h.entryService.CreateEntry(input, userID)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, "entry created", entry)
}

// ListEntries godoc
// @Summary      List financial entries
// @Description  Returns paginated entries. Admins see all; others see only their own.
// @Tags         entries
// @Produce      json
// @Param        page      query     int     false  "Page number (default 1)"
// @Param        page_size query     int     false  "Page size (default 20)"
// @Param        category  query     string  false  "Filter by category"
// @Param        type      query     string  false  "Filter by type: income or expense"
// @Param        date_from query     string  false  "Filter from date (YYYY-MM-DD)"
// @Param        date_to   query     string  false  "Filter to date (YYYY-MM-DD)"
// @Success      200       {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /entries [get]
func (h *EntryHandler) ListEntries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := models.EntryFilter{
		Category: c.Query("category"),
		Type:     models.EntryType(c.Query("type")),
		Page:     page,
		PageSize: pageSize,
	}

	// Parse optional date filters
	if df := c.Query("date_from"); df != "" {
		t, err := time.Parse("2006-01-02", df)
		if err != nil {
			utils.BadRequest(c, "date_from must be in YYYY-MM-DD format")
			return
		}
		filter.DateFrom = &t
	}
	if dt := c.Query("date_to"); dt != "" {
		t, err := time.Parse("2006-01-02", dt)
		if err != nil {
			utils.BadRequest(c, "date_to must be in YYYY-MM-DD format")
			return
		}
		filter.DateTo = &t
	}

	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	entries, total, err := h.entryService.ListEntries(filter, userID, role)
	if err != nil {
		utils.InternalError(c, "failed to list entries")
		return
	}

	utils.Paginated(c, entries, total, page, pageSize)
}

// GetEntry godoc
// @Summary      Get an entry by ID
// @Description  Returns a single entry. Non-admins can only access their own entries.
// @Tags         entries
// @Produce      json
// @Param        id   path      string  true  "Entry ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /entries/{id} [get]
func (h *EntryHandler) GetEntry(c *gin.Context) {
	entryID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	entry, err := h.entryService.GetEntry(entryID, userID, role)
	if err != nil {
		utils.InternalError(c, "failed to get entry")
		return
	}
	if entry == nil {
		utils.NotFound(c, "entry not found")
		return
	}

	utils.OK(c, "entry retrieved", entry)
}

// UpdateEntry godoc
// @Summary      Update a financial entry
// @Description  Partially updates an entry. Only the owner or an admin may update.
// @Tags         entries
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Entry ID"
// @Param        body  body      models.UpdateEntryInput  true  "Fields to update"
// @Success      200   {object}  utils.APIResponse
// @Failure      400   {object}  utils.APIResponse
// @Failure      403   {object}  utils.APIResponse
// @Failure      404   {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /entries/{id} [put]
func (h *EntryHandler) UpdateEntry(c *gin.Context) {
	entryID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	var input models.UpdateEntryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	updated, err := h.entryService.UpdateEntry(entryID, input, userID, role)
	if err != nil {
		if err.Error() == "forbidden" {
			utils.Forbidden(c, "you do not have permission to update this entry")
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}
	if updated == nil {
		utils.NotFound(c, "entry not found")
		return
	}

	utils.OK(c, "entry updated", updated)
}

// DeleteEntry godoc
// @Summary      Delete a financial entry
// @Description  Soft-deletes an entry. Only the owner or an admin may delete.
// @Tags         entries
// @Produce      json
// @Param        id   path      string  true  "Entry ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      403  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /entries/{id} [delete]
func (h *EntryHandler) DeleteEntry(c *gin.Context) {
	entryID := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)
	role := models.Role(c.GetString(middleware.ContextKeyRole))

	err := h.entryService.DeleteEntry(entryID, userID, role)
	if err != nil {
		switch err.Error() {
		case "not found":
			utils.NotFound(c, "entry not found")
		case "forbidden":
			utils.Forbidden(c, "you do not have permission to delete this entry")
		default:
			utils.InternalError(c, "failed to delete entry")
		}
		return
	}

	utils.OK(c, "entry deleted successfully", nil)
}
