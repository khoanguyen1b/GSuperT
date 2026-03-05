package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	topicconversation "gsupert/internal/modules/topic_conversation"
)

type stubProvider struct {
	validateResult   topicconversation.TopicValidationResult
	validateErr      error
	conversation     []topicconversation.DialogueTurn
	conversationErr  error
	validateCalled   bool
	conversationCall bool
	receivedTopic    string
	receivedTurns    int
}

func (s *stubProvider) ValidateTopic(ctx context.Context, rawTopic string) (topicconversation.TopicValidationResult, error) {
	s.validateCalled = true
	return s.validateResult, s.validateErr
}

func (s *stubProvider) GenerateConversation(ctx context.Context, topic string, turnCount int) ([]topicconversation.DialogueTurn, error) {
	s.conversationCall = true
	s.receivedTopic = topic
	s.receivedTurns = turnCount
	return s.conversation, s.conversationErr
}

func TestServiceGenerate_InvalidTurns_ReturnsValidationError(t *testing.T) {
	service := topicconversation.NewService(&stubProvider{})

	_, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "Travel",
		Turns: 10,
	})

	var validationErr *topicconversation.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "turns must be between 20 and 100", validationErr.Fields["turns"])
}

func TestServiceGenerate_InvalidTurnsStep_ReturnsValidationError(t *testing.T) {
	service := topicconversation.NewService(&stubProvider{})

	_, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "Travel",
		Turns: 25,
	})

	var validationErr *topicconversation.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "turns must be a multiple of 10 (e.g. 20, 30, ..., 100)", validationErr.Fields["turns"])
}

func TestServiceGenerate_InvalidTopic_SkipsConversationGeneration(t *testing.T) {
	provider := &stubProvider{
		validateResult: topicconversation.TopicValidationResult{
			IsValid: false,
			Reason:  "Topic rỗng hoặc không có nghĩa tiếng Anh.",
		},
	}
	service := topicconversation.NewService(provider)

	resp, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "@@@",
		Turns: 20,
	})

	assert.NoError(t, err)
	assert.True(t, provider.validateCalled)
	assert.False(t, provider.conversationCall)
	assert.False(t, resp.Valid)
	assert.Equal(t, "Topic rỗng hoặc không có nghĩa tiếng Anh.", resp.ValidationMessage)
	assert.Empty(t, resp.Turns)
}

func TestServiceGenerate_ValidTopic_GeneratesConversation(t *testing.T) {
	provider := &stubProvider{
		validateResult: topicconversation.TopicValidationResult{
			IsValid:         true,
			NormalizedTopic: "Airport travel conversation",
			Reason:          "Topic hợp lệ.",
		},
		conversation: []topicconversation.DialogueTurn{
			{Turn: 1, Speaker: "A", TextEN: "Hello, where are you flying today?", TextVI: "Chào bạn, hôm nay bạn bay đi đâu?"},
			{Turn: 2, Speaker: "B", TextEN: "I am flying to Singapore for work.", TextVI: "Tôi bay đến Singapore để công tác."},
		},
	}
	service := topicconversation.NewService(provider)

	resp, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "airport trip",
		Turns: 20,
	})

	assert.NoError(t, err)
	assert.True(t, provider.validateCalled)
	assert.True(t, provider.conversationCall)
	assert.Equal(t, "Airport travel conversation", provider.receivedTopic)
	assert.Equal(t, 20, provider.receivedTurns)
	assert.True(t, resp.Valid)
	assert.Equal(t, "Topic hợp lệ.", resp.ValidationMessage)
	assert.Equal(t, 2, resp.TurnCount)
	assert.Len(t, resp.Turns, 2)
}

func TestServiceGenerate_DefaultTurnsWhenMissing(t *testing.T) {
	provider := &stubProvider{
		validateResult: topicconversation.TopicValidationResult{
			IsValid:         true,
			NormalizedTopic: "Daily routines",
		},
		conversation: []topicconversation.DialogueTurn{
			{Turn: 1, Speaker: "A", TextEN: "How do you start your day?", TextVI: "Bạn bắt đầu ngày mới như thế nào?"},
		},
	}
	service := topicconversation.NewService(provider)

	_, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "daily life",
	})

	assert.NoError(t, err)
	assert.Equal(t, topicconversation.DefaultTurns, provider.receivedTurns)
}

func TestServiceGenerate_TypedNilProvider_ReturnsValidationError(t *testing.T) {
	var nilProvider *topicconversation.OpenAIProvider
	service := topicconversation.NewService(nilProvider)

	_, err := service.Generate(context.Background(), topicconversation.GenerateRequest{
		Topic: "travel",
		Turns: 20,
	})

	var validationErr *topicconversation.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "gpt provider is not configured on server", validationErr.Fields["topic"])
}
