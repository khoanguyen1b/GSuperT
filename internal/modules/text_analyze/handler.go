package text_analyze

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Analyze(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}

	if err := ValidateAnalyzeRequest(&req); err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			writeValidationError(c, validationErr)
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorBody{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
				Fields:  nil,
			},
		})
		return
	}

	resp, err := h.service.Analyze(c.Request.Context(), req)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			writeValidationError(c, validationErr)
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:    "INTERNAL_ERROR",
				Message: "failed to analyze text",
				Fields:  nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func writeValidationError(c *gin.Context, err *ValidationError) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: err.Message,
			Fields:  err.Fields,
		},
	})
}
