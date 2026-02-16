package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gsupert/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	common.Success(c, http.StatusOK, tokens)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		common.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	common.Success(c, http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if err := h.service.Logout(userID); err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, gin.H{"message": "logged out successfully"})
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.service.CreateUser(req.Email, req.Password, req.FullName, req.Role)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusCreated, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUser(id)
	if err != nil {
		common.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	common.Success(c, http.StatusOK, user)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, users)
}

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.service.UpdateUser(id, req.Email, req.FullName, req.Role)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteUser(id); err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, gin.H{"message": "user deleted"})
}
