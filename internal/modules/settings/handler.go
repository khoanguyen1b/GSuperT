package settings

import (
	"errors"
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

type UpsertSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

type validationErrorBody struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields"`
	} `json:"error"`
}

func (h *Handler) UpsertSettings(c *gin.Context) {
	var req []UpsertSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, "validation failed", map[string]string{
			"body": err.Error(),
		})
		return
	}

	inputs := make([]UpsertInput, 0, len(req))
	for _, item := range req {
		inputs = append(inputs, UpsertInput{
			Key:   item.Key,
			Value: item.Value,
		})
	}

	settings, err := h.service.UpsertMany(inputs)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			writeValidationError(c, validationErr.Message, validationErr.Fields)
			return
		}
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, settings)
}

func (h *Handler) ListSettings(c *gin.Context) {
	settings, err := h.service.List()
	if err != nil {
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, settings)
}

func (h *Handler) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.service.GetByKey(key)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			writeValidationError(c, validationErr.Message, validationErr.Fields)
			return
		}
		if errors.Is(err, ErrSettingNotFound) {
			common.Error(c, http.StatusNotFound, "NOT_FOUND", "Setting not found")
			return
		}
		common.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	common.Success(c, http.StatusOK, setting)
}

func writeValidationError(c *gin.Context, message string, fields map[string]string) {
	resp := validationErrorBody{}
	resp.Error.Code = "VALIDATION_ERROR"
	resp.Error.Message = message
	resp.Error.Fields = fields
	c.JSON(http.StatusBadRequest, resp)
}
