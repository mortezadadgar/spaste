package models

type Snippet struct {
	ID        int
	Text      string
	Lang      string
	LineCount int
	Address   string
	TimeStamp string
}
