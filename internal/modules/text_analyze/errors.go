package text_analyze

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MaxTextLength        = 20000
	DefaultLinkingMode   = "mvp"
	DefaultSyntaxMode    = "mvp"
	GPTSyntaxMode        = "gpt"
	DefaultMaxChunkWords = 12
)

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Fields)
}

func ValidateAnalyzeRequest(req *AnalyzeRequest) error {
	fields := make(map[string]string)
	if req == nil {
		fields["request"] = "request body is required"
		return &ValidationError{
			Message: "invalid request",
			Fields:  fields,
		}
	}

	if strings.TrimSpace(req.Text) == "" {
		fields["text"] = "text is required"
	} else if utf8.RuneCountInString(req.Text) > MaxTextLength {
		fields["text"] = "text exceeds max length 20000"
	}

	if req.Options != nil {
		if req.Options.Linking != nil {
			if req.Options.Linking.Mode != "" && req.Options.Linking.Mode != DefaultLinkingMode {
				fields["options.linking.mode"] = "only 'mvp' is supported"
			}
			if req.Options.Linking.MaxChunkWords < 0 {
				fields["options.linking.max_chunk_words"] = "must be >= 0"
			}
		}

		if req.Options.Syntax != nil {
			if req.Options.Syntax.Mode != "" && req.Options.Syntax.Mode != DefaultSyntaxMode && req.Options.Syntax.Mode != GPTSyntaxMode {
				fields["options.syntax.mode"] = "only 'mvp' and 'gpt' are supported"
			}
		}
	}

	if len(fields) > 0 {
		return &ValidationError{
			Message: "validation failed",
			Fields:  fields,
		}
	}

	return nil
}
