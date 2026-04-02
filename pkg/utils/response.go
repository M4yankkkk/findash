package utils

import "github.com/gin-gonic/gin"

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse wraps list responses with pagination metadata.
type PaginatedResponse struct {
	Items    interface{} `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	HasMore  bool        `json:"has_more"`
}

// OK sends a 200 response with data.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(200, APIResponse{Success: true, Message: message, Data: data})
}

// Created sends a 201 response with the newly created resource.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(201, APIResponse{Success: true, Message: message, Data: data})
}

// BadRequest sends a 400 response for invalid input.
func BadRequest(c *gin.Context, message string) {
	c.JSON(400, APIResponse{Success: false, Error: message})
}

// Unauthorized sends a 401 response.
func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, APIResponse{Success: false, Error: message})
}

// Forbidden sends a 403 response when the user lacks permission.
func Forbidden(c *gin.Context, message string) {
	c.JSON(403, APIResponse{Success: false, Error: message})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	c.JSON(404, APIResponse{Success: false, Error: message})
}

// Conflict sends a 409 response for duplicate resource errors.
func Conflict(c *gin.Context, message string) {
	c.JSON(409, APIResponse{Success: false, Error: message})
}

// InternalError sends a 500 response and logs the real error server-side.
func InternalError(c *gin.Context, message string) {
	c.JSON(500, APIResponse{Success: false, Error: message})
}

// Paginated sends a 200 response with paginated list data.
func Paginated(c *gin.Context, items interface{}, total, page, pageSize int) {
	c.JSON(200, APIResponse{
		Success: true,
		Data: PaginatedResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			HasMore:  (page * pageSize) < total,
		},
	})
}
