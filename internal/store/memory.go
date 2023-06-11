package store

import (
	"github.com/mortezadadgar/spaste/internal/models"
)

type InMemoryStore struct {
	models []models.Snippet
}

// NewInMemoryStore returns a instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (i *InMemoryStore) Create(snippet *models.Snippet) error {
	i.models = append(i.models, models.Snippet{
		ID:        i.getLastID() + 1,
		Text:      snippet.Text,
		Lang:      snippet.Lang,
		LineCount: snippet.LineCount,
		Address:   snippet.Address,
	})

	return nil
}

func (i *InMemoryStore) Get(addr string) (*models.Snippet, error) {
	for _, model := range i.models {
		if model.Address == addr {
			return &model, nil
		}
	}

	return nil, nil
}

func (i *InMemoryStore) getLastID() int {
	if len(i.models) > 0 {
		return i.models[len(i.models)-1].ID
	}
	return 0
}
