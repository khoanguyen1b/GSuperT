package text_analyze

type AnalyzeRequest struct {
	Text    string          `json:"text"`
	Options *AnalyzeOptions `json:"options,omitempty"`
}

type AnalyzeOptions struct {
	Linking *LinkingOptions `json:"linking,omitempty"`
	Syntax  *SyntaxOptions  `json:"syntax,omitempty"`
}

type LinkingOptions struct {
	Mode          string `json:"mode"`
	MaxChunkWords int    `json:"max_chunk_words"`
}

type SyntaxOptions struct {
	Mode string `json:"mode"`
}

type AnalyzeResponse struct {
	Meta          Meta           `json:"meta"`
	LinkingChunks []LinkingChunk `json:"linking_chunks"`
	Syntax        SyntaxResult   `json:"syntax"`
}

type Meta struct {
	Lang       string `json:"lang"`
	TokenCount int    `json:"token_count"`
	Version    string `json:"version"`
}

type LinkingChunk struct {
	ChunkID     string       `json:"chunk_id"`
	Text        string       `json:"text"`
	Tokens      []ChunkToken `json:"tokens"`
	LinkPoints  []LinkPoint  `json:"link_points"`
	FutureAudio FutureAudio  `json:"future_audio"`
}

type ChunkToken struct {
	I int    `json:"i"`
	T string `json:"t"`
}

type LinkPoint struct {
	Type string       `json:"type"`
	From TokenPointer `json:"from"`
	To   TokenPointer `json:"to"`
	Note string       `json:"note"`
}

type TokenPointer struct {
	TokenI    int    `json:"token_i"`
	CharRange [2]int `json:"char_range"`
}

type FutureAudio struct {
	TTSText  string  `json:"tts_text"`
	AudioURL *string `json:"audio_url"`
}

type SyntaxResult struct {
	Sentences []SentenceSyntax `json:"sentences"`
}

type SentenceSyntax struct {
	SentenceID   string        `json:"sentence_id"`
	Text         string        `json:"text"`
	Tokens       []SyntaxToken `json:"tokens"`
	Phrases      []Phrase      `json:"phrases"`
	Dependencies []Dependency  `json:"dependencies"`
}

type SyntaxToken struct {
	I   int    `json:"i"`
	T   string `json:"t"`
	POS string `json:"pos"`
}

type Phrase struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Span [2]int `json:"span"`
}

type Dependency struct {
	Head int    `json:"head"`
	Dep  int    `json:"dep"`
	Rel  string `json:"rel"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
