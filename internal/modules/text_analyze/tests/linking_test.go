package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	textanalyze "gsupert/internal/modules/text_analyze"
)

func TestBuildLinkingChunks_ConsonantVowelHeuristic(t *testing.T) {
	sentences := textanalyze.SplitSentences("I want to go out and take it with us.")
	chunks := textanalyze.BuildLinkingChunks(sentences, 12)

	if assert.Len(t, chunks, 1) {
		points := chunks[0].LinkPoints
		if assert.Len(t, points, 2) {
			assert.Equal(t, "consonant_vowel_link", points[0].Type)
			assert.Equal(t, 4, points[0].From.TokenI) // out
			assert.Equal(t, 5, points[0].To.TokenI)   // and

			assert.Equal(t, "consonant_vowel_link", points[1].Type)
			assert.Equal(t, 8, points[1].From.TokenI) // with
			assert.Equal(t, 9, points[1].To.TokenI)   // us.
		}
	}
}

func TestBuildLinkingChunks_MaxWords(t *testing.T) {
	sentences := textanalyze.SplitSentences("One two three four five six seven eight nine ten eleven twelve thirteen.")
	chunks := textanalyze.BuildLinkingChunks(sentences, 5)

	assert.Len(t, chunks, 3)
	assert.Len(t, chunks[0].Tokens, 5)
	assert.Len(t, chunks[1].Tokens, 5)
	assert.Len(t, chunks[2].Tokens, 3)
}
