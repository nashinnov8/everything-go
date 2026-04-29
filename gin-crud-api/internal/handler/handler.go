package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/gin-crud-api/pkg/errors"
	"github.com/yourusername/gin-crud-api/pkg/logger"
)

// BaseHandler provides common handler functionality
type BaseHandler struct {
	logger logger.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(logger logger.Logger) BaseHandler {
	return BaseHandler{logger: logger}
}

// Response represents API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MetaInfo represents metadata for list responses
type MetaInfo struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Success responds with success
func (h *BaseHandler) Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta responds with success and metadata
func (h *BaseHandler) SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta *MetaInfo) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error responds with error
func (h *BaseHandler) Error(c *gin.Context, err error) {
	appErr := errors.GetError(err)
	c.JSON(appErr.Status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	})
}

// BindAndValidate binds and validates request
func (h *BaseHandler) BindAndValidate(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		h.Error(c, errors.ErrInvalidRequest)
		return false
	}
	return true
}
