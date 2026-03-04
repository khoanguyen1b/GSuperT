package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	textanalyze "gsupert/internal/modules/text_analyze"
)

type stubSyntaxProvider struct {
	called bool
}

func (s *stubSyntaxProvider) Parse(ctx context.Context, text string) ([]textanalyze.SentenceSyntax, error) {
	s.called = true
	return []textanalyze.SentenceSyntax{
		{
			SentenceID: "s1",
			Text:       text,
			Tokens: []textanalyze.SyntaxToken{
				{I: 0, T: "Hello", POS: "INTJ"},
			},
			Phrases:      []textanalyze.Phrase{},
			Dependencies: []textanalyze.Dependency{},
		},
	}, nil
}

func TestValidateAnalyzeRequest_AllowsGPTSyntaxMode(t *testing.T) {
	req := textanalyze.AnalyzeRequest{
		Text: "Hello world.",
		Options: &textanalyze.AnalyzeOptions{
			Syntax: &textanalyze.SyntaxOptions{Mode: textanalyze.GPTSyntaxMode},
		},
	}

	err := textanalyze.ValidateAnalyzeRequest(&req)
	assert.NoError(t, err)
}

func TestService_Analyze_GPTModeWithoutConfiguredProvider_ReturnsValidationError(t *testing.T) {
	service := textanalyze.NewService(textanalyze.NewMockSyntaxProvider())
	req := textanalyze.AnalyzeRequest{
		Text: "Hello world.",
		Options: &textanalyze.AnalyzeOptions{
			Syntax: &textanalyze.SyntaxOptions{Mode: textanalyze.GPTSyntaxMode},
		},
	}

	_, err := service.Analyze(context.Background(), req)
	var validationErr *textanalyze.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "mode 'gpt' is not configured on server", validationErr.Fields["options.syntax.mode"])
}

func TestService_Analyze_GPTModeUsesGPTProvider(t *testing.T) {
	stubGPT := &stubSyntaxProvider{}
	service := textanalyze.NewService(textanalyze.NewMockSyntaxProvider(), stubGPT)
	req := textanalyze.AnalyzeRequest{
		Text: "Hello world.",
		Options: &textanalyze.AnalyzeOptions{
			Syntax: &textanalyze.SyntaxOptions{Mode: textanalyze.GPTSyntaxMode},
		},
	}

	resp, err := service.Analyze(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, stubGPT.called)
	assert.Len(t, resp.Syntax.Sentences, 1)
}
