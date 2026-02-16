package notes

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

type CreateNoteRequest struct {
	Content    string `json:"content" binding:"required"`
	CustomerID string `json:"customer_id" binding:"required"`
}

func (h *Handler) CreateNote(c *gin.Context) {
	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	note, err := h.service.CreateNote(req.Content, req.CustomerID)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusCreated, note)
}

func (h *Handler) GetNote(c *gin.Context) {
	id := c.Param("id")
	note, err := h.service.GetNote(id)
	if err != nil {
		common.Error(c, http.StatusNotFound, "NOT_FOUND", "Note not found")
		return
	}

	common.Success(c, http.StatusOK, note)
}

func (h *Handler) ListNotes(c *gin.Context) {
	customerID := c.Query("customer_id")
	var notes []Note
	var err error

	if customerID != "" {
		notes, err = h.service.ListNotesByCustomer(customerID)
	} else {
		notes, err = h.service.ListNotes()
	}

	if err != nil {
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, notes)
}

type UpdateNoteRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *Handler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	note, err := h.service.UpdateNote(id, req.Content)
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, note)
}

func (h *Handler) DeleteNote(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteNote(id); err != nil {
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, nil)
}
