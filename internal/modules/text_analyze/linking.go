package text_analyze

import (
	"fmt"
	"strings"
)

func BuildLinkingChunks(sentences []string, maxChunkWords int) []LinkingChunk {
	if maxChunkWords <= 0 {
		maxChunkWords = 12
	}

	chunks := make([]LinkingChunk, 0)
	chunkCounter := 1

	for _, sentence := range sentences {
		tokens := TokenizeSentence(sentence)
		if len(tokens) == 0 {
			continue
		}

		for _, part := range splitSentenceIntoChunks(tokens, maxChunkWords) {
			chunkTokens := make([]ChunkToken, 0, len(part))
			for i, token := range part {
				chunkTokens = append(chunkTokens, ChunkToken{I: i, T: token.T})
			}

			chunkText := joinChunkText(part)
			chunks = append(chunks, LinkingChunk{
				ChunkID:    fmt.Sprintf("c%d", chunkCounter),
				Text:       chunkText,
				Tokens:     chunkTokens,
				LinkPoints: buildLinkPoints(part),
				FutureAudio: FutureAudio{
					TTSText:  chunkText,
					AudioURL: nil,
				},
			})
			chunkCounter++
		}
	}

	return chunks
}

func splitSentenceIntoChunks(tokens []Token, maxChunkWords int) [][]Token {
	chunks := make([][]Token, 0)
	start := 0
	for start < len(tokens) {
		end := start + maxChunkWords
		if end >= len(tokens) {
			chunks = append(chunks, tokens[start:])
			break
		}

		preferredEnd := -1
		for i := end - 1; i > start; i-- {
			if hasPreferredBreak(tokens[i].T) {
				preferredEnd = i + 1
				break
			}
		}
		if preferredEnd > 0 {
			end = preferredEnd
		}

		chunks = append(chunks, tokens[start:end])
		start = end
	}

	return chunks
}

func buildLinkPoints(tokens []Token) []LinkPoint {
	points := make([]LinkPoint, 0)
	for i := 0; i < len(tokens)-1; i++ {
		if EndsWithConsonant(tokens[i].Norm) && IsVowelStart(tokens[i+1].Norm) {
			points = append(points, LinkPoint{
				Type: "consonant_vowel_link",
				From: TokenPointer{
					TokenI:    i,
					CharRange: [2]int{0, 0},
				},
				To: TokenPointer{
					TokenI:    i + 1,
					CharRange: [2]int{0, 0},
				},
				Note: "previous token ends with consonant and next token starts with vowel",
			})
		}
	}
	return points
}

func joinChunkText(tokens []Token) string {
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		parts = append(parts, t.T)
	}
	return strings.Join(parts, " ")
}

func hasPreferredBreak(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}

	runes := []rune(token)
	for len(runes) > 0 {
		last := runes[len(runes)-1]
		if last == ',' || last == ';' {
			return true
		}
		if last == '"' || last == '\'' || last == ')' || last == ']' {
			runes = runes[:len(runes)-1]
			continue
		}
		return false
	}
	return false
}
