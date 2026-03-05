package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	topicconversation "gsupert/internal/modules/topic_conversation"
)

type handlerStubProvider struct {
	validateResult topicconversation.TopicValidationResult
	validateErr    error
}

func (s *handlerStubProvider) ValidateTopic(ctx context.Context, rawTopic string) (topicconversation.TopicValidationResult, error) {
	if s.validateErr != nil {
		return topicconversation.TopicValidationResult{}, s.validateErr
	}
	return s.validateResult, nil
}

func (s *handlerStubProvider) GenerateConversation(ctx context.Context, topic string, turnCount int) ([]topicconversation.DialogueTurn, error) {
	return []topicconversation.DialogueTurn{
		{Turn: 1, Speaker: "A", TextEN: "Hello", TextVI: "Xin chao"},
	}, nil
}

func TestHandlerGenerate_ProviderError_ReturnsBadGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &handlerStubProvider{
		validateErr: &topicconversation.ProviderError{
			Message: "GPT API error (401): invalid api key",
		},
	}
	service := topicconversation.NewService(provider)
	handler := topicconversation.NewHandler(service)

	router := gin.New()
	router.POST("/topic-conversation", handler.Generate)

	body := map[string]interface{}{
		"topic": "travel plan",
		"turns": 20,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/topic-conversation", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadGateway, resp.Code)
	assert.Contains(t, resp.Body.String(), "UPSTREAM_ERROR")
}

func TestHandlerGenerate_InvalidTurns_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := &handlerStubProvider{
		validateResult: topicconversation.TopicValidationResult{
			IsValid:         true,
			NormalizedTopic: "Travel plan",
			Reason:          "Topic hợp lệ.",
		},
	}
	service := topicconversation.NewService(provider)
	handler := topicconversation.NewHandler(service)

	router := gin.New()
	router.POST("/topic-conversation", handler.Generate)

	body := map[string]interface{}{
		"topic": "travel plan",
		"turns": 25,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/topic-conversation", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "VALIDATION_ERROR")
	assert.Contains(t, resp.Body.String(), "multiple of 10")
}
