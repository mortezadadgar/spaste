package modules

type Paste struct {
	ID        int    `json:"-"`
	Text      string `json:"text,omitempty"`
	Lang      string `json:"lang,omitempty"`
	LineCount int    `json:"linecount,omitempty"`
	Address   string `json:"address,omitempty"`
	TimeStamp string `json:"-"`
}

type TemplateData struct {
	Address         string
	TextHighlighted string
	LineCount       int
	Lang            string

	Message     string
	IncludeHome bool
}
