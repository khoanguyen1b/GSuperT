package text_analyze

import (
	"strings"
	"unicode"
)

type Token struct {
	I    int
	T    string
	Norm string
}

func SplitSentences(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	sentences := make([]string, 0)
	start := 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '.', '?', '!', '\n':
			part := strings.TrimSpace(text[start : i+1])
			if part != "" {
				sentences = append(sentences, part)
			}
			start = i + 1
		}
	}

	if start < len(text) {
		part := strings.TrimSpace(text[start:])
		if part != "" {
			sentences = append(sentences, part)
		}
	}

	return sentences
}

func TokenizeSentence(sentence string) []Token {
	words := strings.Fields(sentence)
	tokens := make([]Token, 0, len(words))
	for i, w := range words {
		tokens = append(tokens, Token{
			I:    i,
			T:    w,
			Norm: NormalizeToken(w),
		})
	}
	return tokens
}

func NormalizeToken(token string) string {
	token = strings.ToLower(token)
	runes := []rune(token)
	if len(runes) == 0 {
		return ""
	}

	start := 0
	end := len(runes) - 1

	for start <= end && shouldTrimRune(runes[start]) {
		start++
	}
	for end >= start && shouldTrimRune(runes[end]) {
		end--
	}

	if start > end {
		return ""
	}
	return string(runes[start : end+1])
}

func IsVowelStart(tokenNorm string) bool {
	if tokenNorm == "" {
		return false
	}
	first := rune(tokenNorm[0])
	switch first {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

func EndsWithConsonant(tokenNorm string) bool {
	if tokenNorm == "" {
		return false
	}
	last := rune(tokenNorm[len(tokenNorm)-1])
	if !unicode.IsLetter(last) {
		return false
	}
	switch last {
	case 'a', 'e', 'i', 'o', 'u':
		return false
	default:
		return true
	}
}

func shouldTrimRune(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r)
}
