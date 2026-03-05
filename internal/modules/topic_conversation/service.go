package topic_conversation

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Provider interface {
	ValidateTopic(ctx context.Context, rawTopic string) (TopicValidationResult, error)
	GenerateConversation(ctx context.Context, topic string, turnCount int) ([]DialogueTurn, error)
}

type TopicValidationResult struct {
	IsValid         bool
	NormalizedTopic string
	Reason          string
}

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	if isNilProvider(provider) {
		provider = nil
	}
	return &Service{provider: provider}
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if err := ValidateGenerateRequest(&req); err != nil {
		return GenerateResponse{}, err
	}

	if s.provider == nil {
		return GenerateResponse{}, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"topic": "gpt provider is not configured on server",
			},
		}
	}

	turnCount := req.Turns
	if turnCount == 0 {
		turnCount = DefaultTurns
	}

	validationResult, err := s.provider.ValidateTopic(ctx, req.Topic)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("validate topic: %w", err)
	}

	response := GenerateResponse{
		Valid:             validationResult.IsValid,
		Topic:             strings.TrimSpace(validationResult.NormalizedTopic),
		ValidationMessage: strings.TrimSpace(validationResult.Reason),
	}

	if response.ValidationMessage == "" {
		if response.Valid {
			response.ValidationMessage = "Topic hợp lệ."
		} else {
			response.ValidationMessage = "Topic không hợp lệ hoặc chưa có nghĩa tiếng Anh rõ ràng."
		}
	}

	if !response.Valid {
		return response, nil
	}

	if response.Topic == "" {
		response.Topic = strings.TrimSpace(req.Topic)
	}

	dialogueTurns, err := s.provider.GenerateConversation(ctx, response.Topic, turnCount)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("generate conversation: %w", err)
	}

	response.TurnCount = len(dialogueTurns)
	response.Turns = dialogueTurns

	return response, nil
}

func isNilProvider(provider Provider) bool {
	if provider == nil {
		return true
	}

	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return value.IsNil()
	default:
		return false
	}
}
