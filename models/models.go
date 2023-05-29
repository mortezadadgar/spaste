package models

type Snippet struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Address string `json:"address"`
}
