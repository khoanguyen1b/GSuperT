package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	textanalyze "gsupert/internal/modules/text_analyze"
)

func TestSplitSentences(t *testing.T) {
	input := "Hello world.\nThis is great! Are you okay?"
	got := textanalyze.SplitSentences(input)

	assert.Equal(t, []string{
		"Hello world.",
		"This is great!",
		"Are you okay?",
	}, got)
}

func TestNormalizeToken(t *testing.T) {
	assert.Equal(t, "hello", textanalyze.NormalizeToken("\"Hello,\""))
	assert.Equal(t, "world", textanalyze.NormalizeToken("(WORLD)"))
	assert.Equal(t, "", textanalyze.NormalizeToken("..."))
}

func TestTokenHelpers(t *testing.T) {
	assert.True(t, textanalyze.IsVowelStart("apple"))
	assert.False(t, textanalyze.IsVowelStart("banana"))
	assert.True(t, textanalyze.EndsWithConsonant("want"))
	assert.False(t, textanalyze.EndsWithConsonant("go"))
}
