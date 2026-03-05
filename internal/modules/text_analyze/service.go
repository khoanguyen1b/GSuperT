package text_analyze

import (
	"context"
	"reflect"
)

type Service struct {
	syntaxProvider    SyntaxProvider
	gptSyntaxProvider SyntaxProvider
}

func NewService(syntaxProvider SyntaxProvider, gptSyntaxProvider ...SyntaxProvider) *Service {
	if isNilSyntaxProvider(syntaxProvider) {
		syntaxProvider = NewMockSyntaxProvider()
	}

	var gptProvider SyntaxProvider
	if len(gptSyntaxProvider) > 0 {
		gptProvider = gptSyntaxProvider[0]
	}
	if isNilSyntaxProvider(gptProvider) {
		gptProvider = nil
	}

	return &Service{
		syntaxProvider:    syntaxProvider,
		gptSyntaxProvider: gptProvider,
	}
}

func (s *Service) Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error) {
	options := resolveOptions(req.Options)
	sentences := SplitSentences(req.Text)

	tokenCount := 0
	for _, sentence := range sentences {
		tokenCount += len(TokenizeSentence(sentence))
	}

	linkingChunks := BuildLinkingChunks(sentences, options.maxChunkWords)

	syntaxProvider, err := s.resolveSyntaxProvider(options.syntaxMode)
	if err != nil {
		return AnalyzeResponse{}, err
	}

	syntaxSentences, err := syntaxProvider.Parse(ctx, req.Text)
	if err != nil {
		return AnalyzeResponse{}, err
	}

	return AnalyzeResponse{
		Meta: Meta{
			Lang:       "en",
			TokenCount: tokenCount,
			Version:    "mvp-1",
		},
		LinkingChunks: linkingChunks,
		Syntax: SyntaxResult{
			Sentences: syntaxSentences,
		},
	}, nil
}

type resolvedOptions struct {
	linkingMode   string
	maxChunkWords int
	syntaxMode    string
}

func resolveOptions(opts *AnalyzeOptions) resolvedOptions {
	out := resolvedOptions{
		linkingMode:   DefaultLinkingMode,
		maxChunkWords: DefaultMaxChunkWords,
		syntaxMode:    DefaultSyntaxMode,
	}

	if opts == nil {
		return out
	}

	if opts.Linking != nil {
		if opts.Linking.Mode != "" {
			out.linkingMode = opts.Linking.Mode
		}
		if opts.Linking.MaxChunkWords > 0 {
			out.maxChunkWords = opts.Linking.MaxChunkWords
		}
	}

	if opts.Syntax != nil && opts.Syntax.Mode != "" {
		out.syntaxMode = opts.Syntax.Mode
	}

	return out
}

func (s *Service) resolveSyntaxProvider(mode string) (SyntaxProvider, error) {
	if mode == GPTSyntaxMode {
		if s.gptSyntaxProvider == nil {
			return nil, &ValidationError{
				Message: "validation failed",
				Fields: map[string]string{
					"options.syntax.mode": "mode 'gpt' is not configured on server",
				},
			}
		}
		return s.gptSyntaxProvider, nil
	}

	return s.syntaxProvider, nil
}

func isNilSyntaxProvider(provider SyntaxProvider) bool {
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
