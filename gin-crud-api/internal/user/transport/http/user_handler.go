package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/gin-crud-api/internal/handler"
	"github.com/yourusername/gin-crud-api/internal/user/domain"
	"github.com/yourusername/gin-crud-api/internal/user/service"
	"github.com/yourusername/gin-crud-api/pkg/errors"
	"github.com/yourusername/gin-crud-api/pkg/logger"
)

type UserHandler struct {
	handler.BaseHandler
	userService service.UserService
}

func NewUserHandler(userService service.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		BaseHandler: handler.NewBaseHandler(logger),
		userService: userService,
	}
}

// Create godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.CreateUserRequest true "User creation request"
// @Success 201 {object} handler.Response{data=domain.UserResponse}
// @Failure 400 {object} handler.Response
// @Failure 409 {object} handler.Response
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req domain.CreateUserRequest
	if !h.BindAndValidate(c, &req) {
		return
	}

	user, err := h.userService.Create(c.Request.Context(), req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nethttp.StatusCreated, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get a user by their UUID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} handler.Response{data=domain.UserResponse}
// @Failure 404 {object} handler.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.Error(c, errors.ErrInvalidUserId)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nethttp.StatusOK, user)
}

// List godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} handler.Response{data=[]domain.UserResponse,meta=handler.MetaInfo}
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	page := c.GetInt("page")
	if page == 0 {
		page = 1
	}
	pageSize := c.GetInt("page_size")
	if pageSize == 0 {
		pageSize = 10
	}

	result, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.SuccessWithMeta(c, nethttp.StatusOK, result.Users, &handler.MetaInfo{
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// Update godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body domain.UpdateUserRequest true "User update request"
// @Success 200 {object} handler.Response{data=domain.UserResponse}
// @Failure 400 {object} handler.Response
// @Failure 404 {object} handler.Response
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.Error(c, errors.ErrInvalidUserId)
		return
	}

	var req domain.UpdateUserRequest
	if !h.BindAndValidate(c, &req) {
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, req)
	if err != nil {
		h.Error(c, err)
		return
	}

	h.Success(c, nethttp.StatusOK, user)
}

// Delete godoc
// @Summary Delete user
// @Description Soft delete a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} handler.Response
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.Error(c, errors.ErrInvalidUserId)
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		h.Error(c, err)
		return
	}

	c.Status(nethttp.StatusNoContent)
}
