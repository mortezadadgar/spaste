package models

type Snippet struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Lang    string `json:"lang"`
	Address string `json:"address"`
}
