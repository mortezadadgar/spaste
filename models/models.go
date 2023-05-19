package models

type Snippet struct {
	Id      int    `json:"id"`
	Text    string `json:"text"`
	Address string `json:"address"`
}
