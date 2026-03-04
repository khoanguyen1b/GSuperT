package text_analyze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultOpenAIModel   = "gpt-4.1-mini"
)

type OpenAISyntaxProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAISyntaxProvider(apiKey, model, baseURL string, httpClient *http.Client) *OpenAISyntaxProvider {
	if strings.TrimSpace(apiKey) == "" {
		return nil
	}

	if strings.TrimSpace(model) == "" {
		model = defaultOpenAIModel
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOpenAIBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 25 * time.Second}
	}

	return &OpenAISyntaxProvider{
		apiKey:     apiKey,
		model:      model,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (p *OpenAISyntaxProvider) Parse(ctx context.Context, text string) ([]SentenceSyntax, error) {
	if strings.TrimSpace(text) == "" {
		return []SentenceSyntax{}, nil
	}

	reqBody := openAIChatCompletionRequest{
		Model:       p.model,
		Temperature: 0,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: strings.TrimSpace(`
You are an English syntax parser.
Return strict JSON only with this shape:
{
  "sentences": [
    {
      "sentence_id": "s1",
      "text": "original sentence text",
      "tokens": [{"i":0,"t":"I","pos":"PRON"}],
      "phrases": [{"id":"p1","type":"NP","span":[0,0]}],
      "dependencies": [{"head":1,"dep":0,"rel":"nsubj"}]
    }
  ]
}
Rules:
- Keep token order as it appears in each sentence.
- Use uppercase Universal POS tags (e.g. NOUN, VERB, ADJ, ADV, PRON, DET, ADP, AUX, PART, CCONJ, SCONJ, NUM, PROPN, INTJ, PUNCT, X).
- Span is inclusive token indexes [start, end].
- Use 0-based token indexes for dependencies. For root token use head = -1 and rel = "root".
- Return JSON only, no markdown fences, no extra keys.
`),
			},
			{
				Role:    "user",
				Content: text,
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create openai request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call openai api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("openai api error (%d): %s", resp.StatusCode, extractOpenAIErrorMessage(body))
	}

	var completion openAIChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("openai api returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("openai api returned empty content")
	}

	parsedJSON := extractJSONBody(content)
	if parsedJSON == "" {
		return nil, fmt.Errorf("openai output does not contain valid json object")
	}

	var parsed openAISyntaxPayload
	if err := json.Unmarshal([]byte(parsedJSON), &parsed); err != nil {
		return nil, fmt.Errorf("decode syntax json: %w", err)
	}

	return normalizeParsedSyntax(parsed, text)
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

type openAISyntaxPayload struct {
	Sentences []openAISentence `json:"sentences"`
}

type openAISentence struct {
	SentenceID   string              `json:"sentence_id"`
	Text         string              `json:"text"`
	Tokens       []openAISyntaxToken `json:"tokens"`
	Phrases      []openAIPhrase      `json:"phrases"`
	Dependencies []Dependency        `json:"dependencies"`
}

type openAISyntaxToken struct {
	I   int    `json:"i"`
	T   string `json:"t"`
	POS string `json:"pos"`
}

type openAIPhrase struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Span []int  `json:"span"`
}

func normalizeParsedSyntax(parsed openAISyntaxPayload, originalText string) ([]SentenceSyntax, error) {
	if len(parsed.Sentences) == 0 {
		return nil, fmt.Errorf("missing sentences in syntax output")
	}

	fallbackSentenceTexts := SplitSentences(originalText)
	out := make([]SentenceSyntax, 0, len(parsed.Sentences))

	for sentenceIndex, sentence := range parsed.Sentences {
		sentenceID := strings.TrimSpace(sentence.SentenceID)
		if sentenceID == "" {
			sentenceID = fmt.Sprintf("s%d", sentenceIndex+1)
		}

		sentenceText := strings.TrimSpace(sentence.Text)
		if sentenceText == "" && sentenceIndex < len(fallbackSentenceTexts) {
			sentenceText = fallbackSentenceTexts[sentenceIndex]
		}

		tokens := make([]SyntaxToken, 0, len(sentence.Tokens))
		for tokenIndex, token := range sentence.Tokens {
			tokenText := strings.TrimSpace(token.T)
			if tokenText == "" {
				continue
			}
			pos := strings.ToUpper(strings.TrimSpace(token.POS))
			if pos == "" {
				pos = "X"
			}
			tokens = append(tokens, SyntaxToken{
				I:   tokenIndex,
				T:   tokenText,
				POS: pos,
			})
		}

		if len(tokens) == 0 && sentenceText != "" {
			tokenized := TokenizeSentence(sentenceText)
			for tokenIndex, token := range tokenized {
				tokens = append(tokens, SyntaxToken{
					I:   tokenIndex,
					T:   token.T,
					POS: "X",
				})
			}
		}

		phrases := make([]Phrase, 0, len(sentence.Phrases))
		for phraseIndex, phrase := range sentence.Phrases {
			if len(phrase.Span) != 2 {
				continue
			}
			start := phrase.Span[0]
			end := phrase.Span[1]
			if start < 0 || end < start || end >= len(tokens) {
				continue
			}

			phraseID := strings.TrimSpace(phrase.ID)
			if phraseID == "" {
				phraseID = fmt.Sprintf("p%d", phraseIndex+1)
			}
			phraseType := strings.TrimSpace(phrase.Type)
			if phraseType == "" {
				phraseType = "UNK"
			}

			phrases = append(phrases, Phrase{
				ID:   phraseID,
				Type: phraseType,
				Span: [2]int{start, end},
			})
		}

		dependencies := make([]Dependency, 0, len(sentence.Dependencies))
		for _, dep := range sentence.Dependencies {
			if dep.Dep < 0 || dep.Dep >= len(tokens) {
				continue
			}
			if dep.Head != -1 && (dep.Head < 0 || dep.Head >= len(tokens)) {
				continue
			}
			rel := strings.TrimSpace(dep.Rel)
			if rel == "" {
				rel = "dep"
			}

			dependencies = append(dependencies, Dependency{
				Head: dep.Head,
				Dep:  dep.Dep,
				Rel:  rel,
			})
		}

		out = append(out, SentenceSyntax{
			SentenceID:   sentenceID,
			Text:         sentenceText,
			Tokens:       tokens,
			Phrases:      phrases,
			Dependencies: dependencies,
		})
	}

	return out, nil
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
	if err := json.Unmarshal(body, &parsed); err == nil && strings.TrimSpace(parsed.Error.Message) != "" {
		return parsed.Error.Message
	}

	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return "unknown error"
	}
	if len(raw) > 300 {
		return raw[:300]
	}
	return raw
}
