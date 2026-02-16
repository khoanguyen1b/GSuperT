package customers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/gsupert/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	customer, err := h.service.CreateCustomer(req.Name, req.Email, req.Phone, req.Address)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusCreated, customer)
}

func (h *Handler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	customer, err := h.service.GetCustomer(id)
	if err != nil {
		common.Error(c, http.StatusNotFound, "NOT_FOUND", "Customer not found")
		return
	}

	common.Success(c, http.StatusOK, customer)
}

func (h *Handler) ListCustomers(c *gin.Context) {
	customers, err := h.service.ListCustomers()
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, customers)
}

type UpdateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	customer, err := h.service.UpdateCustomer(id, req.Name, req.Email, req.Phone, req.Address)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, customer)
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteCustomer(id); err != nil {
		common.Error(c, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, gin.H{"message": "customer deleted"})
}
