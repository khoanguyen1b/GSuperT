package topic_conversation

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[topic-conversation] invalid request body: %v", err)
		writeValidationError(c, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"body": err.Error(),
			},
		})
		return
	}
	log.Printf(
		"[topic-conversation] request received | topic=%q | turns=%d",
		previewTopic(req.Topic),
		req.Turns,
	)

	resp, err := h.service.Generate(c.Request.Context(), req)
	if err != nil {
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			log.Printf("[topic-conversation] validation error: %+v", validationErr.Fields)
			writeValidationError(c, validationErr)
			return
		}

		var providerErr *ProviderError
		if errors.As(err, &providerErr) {
			log.Printf("[topic-conversation] upstream error: %v", providerErr.Message)
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error: ErrorBody{
					Code:    "UPSTREAM_ERROR",
					Message: providerErr.Message,
				},
			})
			return
		}

		log.Printf("[topic-conversation] internal error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:    "INTERNAL_ERROR",
				Message: "failed to generate topic conversation",
			},
		})
		return
	}

	log.Printf(
		"[topic-conversation] success | valid=%t | normalized_topic=%q | turn_count=%d",
		resp.Valid,
		previewTopic(resp.Topic),
		resp.TurnCount,
	)
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

func previewTopic(topic string) string {
	trimmed := strings.TrimSpace(topic)
	if trimmed == "" {
		return ""
	}

	const maxChars = 120
	runes := []rune(trimmed)
	if len(runes) <= maxChars {
		return trimmed
	}

	return string(runes[:maxChars]) + "..."
}
