package text_analyze

import "context"

type Service struct {
	syntaxProvider SyntaxProvider
}

func NewService(syntaxProvider SyntaxProvider) *Service {
	if syntaxProvider == nil {
		syntaxProvider = NewMockSyntaxProvider()
	}
	return &Service{syntaxProvider: syntaxProvider}
}

func (s *Service) Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error) {
	options := resolveOptions(req.Options)
	sentences := SplitSentences(req.Text)

	tokenCount := 0
	for _, sentence := range sentences {
		tokenCount += len(TokenizeSentence(sentence))
	}

	linkingChunks := BuildLinkingChunks(sentences, options.maxChunkWords)

	syntaxSentences, err := s.syntaxProvider.Parse(ctx, req.Text)
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
