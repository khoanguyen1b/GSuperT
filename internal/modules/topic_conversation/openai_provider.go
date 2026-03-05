package topic_conversation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4.1-mini"
	maxConversationRetry = 3
)

type OpenAIProvider struct {
	apiKey         string
	apiKeyResolver func(context.Context) (string, error)
	model          string
	baseURL        string
	httpClient     *http.Client
}

func NewOpenAIProvider(
	apiKey, model, baseURL string,
	httpClient *http.Client,
	apiKeyResolver ...func(context.Context) (string, error),
) *OpenAIProvider {
	var resolver func(context.Context) (string, error)
	if len(apiKeyResolver) > 0 {
		resolver = apiKeyResolver[0]
	}

	if strings.TrimSpace(apiKey) == "" && resolver == nil {
		return nil
	}

	if strings.TrimSpace(model) == "" {
		model = defaultOpenAIModel
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOpenAIBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &OpenAIProvider{
		apiKey:         strings.TrimSpace(apiKey),
		apiKeyResolver: resolver,
		model:          model,
		baseURL:        strings.TrimRight(baseURL, "/"),
		httpClient:     httpClient,
	}
}

func (p *OpenAIProvider) ValidateTopic(ctx context.Context, rawTopic string) (TopicValidationResult, error) {
	log.Printf("[topic-conversation] validate topic | topic=%q", limitForLog(rawTopic, 120))

	requestBody := openAIChatCompletionRequest{
		Model:       p.model,
		Temperature: 0,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: strings.TrimSpace(`
You validate user input for an English speaking topic.
Return strict JSON only with this shape:
{
  "is_valid": true,
  "normalized_topic": "topic in concise English",
  "reason": "Vietnamese message for UI"
}
Rules:
- is_valid=false when input is empty, meaningless, random symbols, or not a meaningful English topic.
- If is_valid=true, normalized_topic must be a short and clear English topic title.
- reason must always be in Vietnamese.
- Return JSON only, no markdown fences, no extra keys.
`),
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Topic input: %q", rawTopic),
			},
		},
	}

	jsonBody, err := p.chatCompletion(ctx, requestBody)
	if err != nil {
		return TopicValidationResult{}, err
	}

	var parsed topicValidationPayload
	if err := json.Unmarshal([]byte(jsonBody), &parsed); err != nil {
		log.Printf(
			"[topic-conversation] validate topic parse error | json=%q | err=%v",
			limitForLog(jsonBody, 400),
			err,
		)
		return TopicValidationResult{}, &ProviderError{
			Message: "GPT trả về dữ liệu validate topic không đúng định dạng. Vui lòng thử lại.",
		}
	}

	result := TopicValidationResult{
		IsValid:         parsed.IsValid,
		NormalizedTopic: strings.TrimSpace(parsed.NormalizedTopic),
		Reason:          strings.TrimSpace(parsed.Reason),
	}

	if result.IsValid && result.NormalizedTopic == "" {
		result.NormalizedTopic = strings.TrimSpace(rawTopic)
	}

	if result.Reason == "" {
		if result.IsValid {
			result.Reason = "Topic hợp lệ."
		} else {
			result.Reason = "Topic rỗng hoặc chưa phải topic tiếng Anh có nghĩa."
		}
	}

	return result, nil
}

func (p *OpenAIProvider) GenerateConversation(ctx context.Context, topic string, turnCount int) ([]DialogueTurn, error) {
	log.Printf(
		"[topic-conversation] generate conversation start | topic=%q | turns=%d",
		limitForLog(topic, 120),
		turnCount,
	)

	if turnCount < MinTurns || turnCount > MaxTurns {
		return nil, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"turns": fmt.Sprintf("turns must be between %d and %d", MinTurns, MaxTurns),
			},
		}
	}
	if turnCount%TurnStep != 0 {
		return nil, &ValidationError{
			Message: "validation failed",
			Fields: map[string]string{
				"turns": fmt.Sprintf("turns must be a multiple of %d (e.g. 20, 30, ..., 100)", TurnStep),
			},
		}
	}

	var lastErr error
	for attempt := 1; attempt <= maxConversationRetry; attempt++ {
		requestBody := openAIChatCompletionRequest{
			Model:       p.model,
			Temperature: 0.4,
			ResponseFormat: map[string]string{
				"type": "json_object",
			},
			Messages: []openAIMessage{
				{
					Role: "system",
					Content: strings.TrimSpace(`
You generate English learning dialogues.
Return strict JSON only.
`),
				},
				{
					Role:    "user",
					Content: buildConversationPrompt(topic, turnCount, attempt),
				},
			},
		}

		jsonBody, err := p.chatCompletion(ctx, requestBody)
		if err != nil {
			log.Printf(
				"[topic-conversation] generate attempt %d/%d failed on chat completion: %v",
				attempt,
				maxConversationRetry,
				err,
			)
			lastErr = err
			continue
		}

		var parsed conversationPayload
		if err := json.Unmarshal([]byte(jsonBody), &parsed); err != nil {
			log.Printf(
				"[topic-conversation] generate attempt %d/%d parse error | json=%q | err=%v",
				attempt,
				maxConversationRetry,
				limitForLog(jsonBody, 400),
				err,
			)
			lastErr = &ProviderError{
				Message: "GPT trả về dữ liệu hội thoại không đúng định dạng JSON. Vui lòng thử lại.",
			}
			continue
		}

		normalized, err := normalizeDialogueTurns(parsed.Turns, turnCount)
		if err != nil {
			log.Printf(
				"[topic-conversation] generate attempt %d/%d normalize error: %v",
				attempt,
				maxConversationRetry,
				err,
			)
			lastErr = &ProviderError{
				Message: fmt.Sprintf("GPT chưa tạo đủ %d lượt hội thoại hợp lệ. Vui lòng thử lại.", turnCount),
			}
			continue
		}

		log.Printf(
			"[topic-conversation] generate conversation success | attempt=%d | returned_turns=%d",
			attempt,
			len(normalized),
		)
		return normalized, nil
	}

	if lastErr == nil {
		lastErr = &ProviderError{
			Message: "Không thể tạo hội thoại từ GPT lúc này. Vui lòng thử lại.",
		}
	}

	return nil, lastErr
}

func (p *OpenAIProvider) chatCompletion(ctx context.Context, requestBody openAIChatCompletionRequest) (string, error) {
	jsonBody, err := p.chatCompletionOnce(ctx, requestBody)
	if err == nil {
		return jsonBody, nil
	}

	if requestBody.ResponseFormat != nil && isUnsupportedResponseFormatError(err) {
		log.Printf("[topic-conversation] response_format unsupported, retrying without response_format")
		fallbackReq := requestBody
		fallbackReq.ResponseFormat = nil
		return p.chatCompletionOnce(ctx, fallbackReq)
	}

	return "", err
}

func (p *OpenAIProvider) chatCompletionOnce(ctx context.Context, requestBody openAIChatCompletionRequest) (string, error) {
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", &ProviderError{
			Message: "Không thể tạo request GPT. Vui lòng thử lại.",
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", &ProviderError{
			Message: "Không thể tạo kết nối đến GPT API.",
		}
	}

	apiKey, err := p.resolveAPIKey(ctx)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", &ProviderError{
			Message: "Không thể gọi GPT API. Vui lòng kiểm tra kết nối và API key.",
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", &ProviderError{
			Message: "Không đọc được phản hồi từ GPT API.",
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Printf(
			"[topic-conversation] gpt api status error | status=%d | body=%q",
			resp.StatusCode,
			limitForLog(string(body), 400),
		)
		return "", &ProviderError{
			Message: fmt.Sprintf("GPT API error (%d): %s", resp.StatusCode, extractOpenAIErrorMessage(body)),
		}
	}

	var completion openAIChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", &ProviderError{
			Message: "Phản hồi GPT API không đúng định dạng.",
		}
	}
	if len(completion.Choices) == 0 {
		return "", &ProviderError{
			Message: "GPT API không trả về nội dung hợp lệ.",
		}
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	if content == "" {
		return "", &ProviderError{
			Message: "GPT API trả về nội dung rỗng.",
		}
	}

	jsonBody := extractJSONBody(content)
	if jsonBody == "" {
		return "", &ProviderError{
			Message: "GPT không trả về JSON hợp lệ.",
		}
	}

	return jsonBody, nil
}

func buildConversationPrompt(topic string, turnCount int, attempt int) string {
	basePrompt := fmt.Sprintf(strings.TrimSpace(`
Create a dialogue about topic: %q

Output JSON exactly with this shape:
{
  "turns": [
    {
      "turn": 1,
      "speaker": "A",
      "text_en": "English line",
      "text_vi": "Vietnamese translation"
    }
  ]
}

Rules:
- Exactly %d turns.
- Speakers must alternate A and B, starting with A.
- text_en must be concise (max 12 English words each turn).
- text_vi is faithful Vietnamese translation for the same turn.
- Keep the conversation coherent from start to finish.
- Return JSON only, no markdown fences, no extra keys.
`), topic, turnCount)

	if attempt <= 1 {
		return basePrompt
	}

	return basePrompt + "\n\nPrevious output was invalid. Regenerate full JSON from scratch and strictly satisfy all rules."
}

func isUnsupportedResponseFormatError(err error) bool {
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		return false
	}
	return isUnsupportedResponseFormatMessage(providerErr.Message)
}

func isUnsupportedResponseFormatMessage(message string) bool {
	lowerMessage := strings.ToLower(strings.TrimSpace(message))
	if lowerMessage == "" {
		return false
	}

	if !strings.Contains(lowerMessage, "response_format") && !strings.Contains(lowerMessage, "json_object") {
		return false
	}

	return strings.Contains(lowerMessage, "not support") ||
		strings.Contains(lowerMessage, "unsupported") ||
		strings.Contains(lowerMessage, "invalid")
}

func normalizeDialogueTurns(turns []openAIConversationTurn, turnCount int) ([]DialogueTurn, error) {
	if turnCount < MinTurns || turnCount > MaxTurns {
		return nil, fmt.Errorf("turn count must be between %d and %d", MinTurns, MaxTurns)
	}
	if len(turns) < turnCount {
		return nil, fmt.Errorf("conversation has %d turns, expected %d", len(turns), turnCount)
	}

	normalized := make([]DialogueTurn, 0, turnCount)
	for i := 0; i < turnCount; i++ {
		englishLine := strings.TrimSpace(turns[i].TextEN)
		vietnameseLine := strings.TrimSpace(turns[i].TextVI)
		if englishLine == "" || vietnameseLine == "" {
			return nil, fmt.Errorf("turn %d is missing text_en or text_vi", i+1)
		}

		speaker := "A"
		if i%2 == 1 {
			speaker = "B"
		}

		normalized = append(normalized, DialogueTurn{
			Turn:    i + 1,
			Speaker: speaker,
			TextEN:  englishLine,
			TextVI:  vietnameseLine,
		})
	}

	return normalized, nil
}

func extractJSONBody(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || start > end {
		return ""
	}

	return trimmed[start : end+1]
}

func extractOpenAIErrorMessage(body []byte) string {
	type apiErrorBody struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
			return msg
		}
	}

	return string(body)
}

func limitForLog(raw string, max int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}

	return string(runes[:max]) + "..."
}

func (p *OpenAIProvider) resolveAPIKey(ctx context.Context) (string, error) {
	if key := strings.TrimSpace(p.apiKey); key != "" {
		return key, nil
	}

	if p.apiKeyResolver != nil {
		resolvedKey, err := p.apiKeyResolver(ctx)
		if err != nil {
			return "", &ProviderError{
				Message: "Không lấy được GPT API key từ app settings.",
			}
		}

		if key := strings.TrimSpace(resolvedKey); key != "" {
			return key, nil
		}
	}

	return "", &ProviderError{
		Message: "GPT API key chưa được cấu hình. Vui lòng cập nhật ở Settings hoặc .env.",
	}
}

type openAIChatCompletionRequest struct {
	Model          string            `json:"model"`
	Temperature    float64           `json:"temperature"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
	Messages       []openAIMessage   `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type topicValidationPayload struct {
	IsValid         bool   `json:"is_valid"`
	NormalizedTopic string `json:"normalized_topic"`
	Reason          string `json:"reason"`
}

type conversationPayload struct {
	Turns []openAIConversationTurn `json:"turns"`
}

type openAIConversationTurn struct {
	Turn    int    `json:"turn"`
	Speaker string `json:"speaker"`
	TextEN  string `json:"text_en"`
	TextVI  string `json:"text_vi"`
}
