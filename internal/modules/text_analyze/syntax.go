package text_analyze

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type SyntaxProvider interface {
	Parse(ctx context.Context, text string) ([]SentenceSyntax, error)
}

type MockSyntaxProvider struct{}

func NewMockSyntaxProvider() *MockSyntaxProvider {
	return &MockSyntaxProvider{}
}

func (m *MockSyntaxProvider) Parse(ctx context.Context, text string) ([]SentenceSyntax, error) {
	sentences := SplitSentences(text)
	out := make([]SentenceSyntax, 0, len(sentences))

	for i, sentenceText := range sentences {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		tokens := TokenizeSentence(sentenceText)
		syntaxTokens := make([]SyntaxToken, 0, len(tokens))
		posTags := make([]string, 0, len(tokens))

		for j, token := range tokens {
			pos := detectPOS(token.Norm)
			posTags = append(posTags, pos)
			syntaxTokens = append(syntaxTokens, SyntaxToken{
				I:   j,
				T:   token.T,
				POS: pos,
			})
		}

		phrases := buildPhrases(posTags)
		for p := range phrases {
			phrases[p].ID = fmt.Sprintf("p%d", p+1)
		}

		out = append(out, SentenceSyntax{
			SentenceID:   fmt.Sprintf("s%d", i+1),
			Text:         sentenceText,
			Tokens:       syntaxTokens,
			Phrases:      phrases,
			Dependencies: make([]Dependency, 0),
		})
	}

	return out, nil
}

func detectPOS(norm string) string {
	switch {
	case pronouns[norm]:
		return "PRON"
	case determiners[norm]:
		return "DET"
	case adpositions[norm]:
		return "ADP"
	case conjunctions[norm]:
		return "CONJ"
	case verbs[norm]:
		return "VERB"
	case strings.HasSuffix(norm, "ed"), strings.HasSuffix(norm, "ing"):
		return "VERB"
	case strings.HasSuffix(norm, "ly"):
		return "ADV"
	case strings.HasSuffix(norm, "ful"),
		strings.HasSuffix(norm, "ous"),
		strings.HasSuffix(norm, "able"),
		strings.HasSuffix(norm, "al"),
		strings.HasSuffix(norm, "ive"):
		return "ADJ"
	default:
		return "NOUN"
	}
}

func buildPhrases(posTags []string) []Phrase {
	type span struct {
		kind string
		from int
		to   int
	}

	spans := make([]span, 0)

	for i := 0; i < len(posTags); {
		_, end, ok := npSpanAt(i, posTags)
		if ok {
			spans = append(spans, span{kind: "NP", from: i, to: end})
			i = end + 1
			continue
		}
		i++
	}

	for i := 0; i < len(posTags); i++ {
		if posTags[i] != "ADP" {
			continue
		}
		_, end, ok := npSpanAt(i+1, posTags)
		if ok {
			spans = append(spans, span{kind: "PP", from: i, to: end})
		}
	}

	for i := 0; i < len(posTags); i++ {
		if posTags[i] != "VERB" {
			continue
		}

		end := i
		next := i + 1
		if next < len(posTags) {
			if posTags[next] == "PRON" {
				end = next
				next++
			} else if _, npEnd, ok := npSpanAt(next, posTags); ok {
				end = npEnd
				next = npEnd + 1
			}
		}

		for next < len(posTags) && posTags[next] == "ADP" {
			_, ppEnd, ok := npSpanAt(next+1, posTags)
			if !ok {
				break
			}
			end = ppEnd
			next = ppEnd + 1
		}

		spans = append(spans, span{kind: "VP", from: i, to: end})
	}

	sort.Slice(spans, func(i, j int) bool {
		if spans[i].from == spans[j].from {
			if spans[i].to == spans[j].to {
				return spans[i].kind < spans[j].kind
			}
			return spans[i].to < spans[j].to
		}
		return spans[i].from < spans[j].from
	})

	phrases := make([]Phrase, 0, len(spans))
	for _, s := range spans {
		phrases = append(phrases, Phrase{
			Type: s.kind,
			Span: [2]int{s.from, s.to},
		})
	}
	return phrases
}

func npSpanAt(start int, posTags []string) (int, int, bool) {
	if start < 0 || start >= len(posTags) {
		return 0, 0, false
	}

	i := start
	if posTags[i] == "DET" {
		i++
	}
	for i < len(posTags) && posTags[i] == "ADJ" {
		i++
	}

	if i >= len(posTags) || posTags[i] != "NOUN" {
		return 0, 0, false
	}

	end := i
	for end+1 < len(posTags) && posTags[end+1] == "NOUN" {
		end++
	}
	return start, end, true
}

var pronouns = map[string]bool{
	"i": true, "you": true, "he": true, "she": true, "it": true,
	"we": true, "they": true, "me": true, "him": true, "her": true,
	"us": true, "them": true,
}

var determiners = map[string]bool{
	"a": true, "an": true, "the": true, "this": true, "that": true,
	"these": true, "those": true, "my": true, "your": true, "his": true,
	"her": true, "our": true, "their": true,
}

var adpositions = map[string]bool{
	"in": true, "on": true, "at": true, "for": true, "to": true,
	"from": true, "with": true, "about": true, "of": true,
}

var conjunctions = map[string]bool{
	"and": true, "or": true, "but": true,
}

var verbs = map[string]bool{
	"be": true, "am": true, "is": true, "are": true, "was": true, "were": true,
	"have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true,
	"go": true, "want": true, "need": true, "make": true, "ask": true, "work": true,
}
