package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	textanalyze "gsupert/internal/modules/text_analyze"
)

func TestMockSyntaxProvider_POSAndPhrases(t *testing.T) {
	provider := textanalyze.NewMockSyntaxProvider()
	sentences, err := provider.Parse(context.Background(), "I want the beautiful car in the city.")
	assert.NoError(t, err)

	if assert.Len(t, sentences, 1) {
		tokens := sentences[0].Tokens
		if assert.Len(t, tokens, 8) {
			assert.Equal(t, "PRON", tokens[0].POS) // I
			assert.Equal(t, "VERB", tokens[1].POS) // want
			assert.Equal(t, "DET", tokens[2].POS)  // the
			assert.Equal(t, "ADJ", tokens[3].POS)  // beautiful
			assert.Equal(t, "NOUN", tokens[4].POS) // car
			assert.Equal(t, "ADP", tokens[5].POS)  // in
			assert.Equal(t, "DET", tokens[6].POS)  // the
			assert.Equal(t, "NOUN", tokens[7].POS) // city.
		}

		assert.NotEmpty(t, sentences[0].Phrases)
		assert.NotNil(t, sentences[0].Dependencies)

		for _, phrase := range sentences[0].Phrases {
			assert.NotEmpty(t, phrase.Type)
			assert.GreaterOrEqual(t, phrase.Span[0], 0)
			assert.GreaterOrEqual(t, phrase.Span[1], phrase.Span[0])
			assert.Less(t, phrase.Span[1], len(tokens))
		}
	}
}
