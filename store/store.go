package store

import (
	"github.com/mortezadadgar/spaste/models"
)

// TODO: implementations
//   - InMemory
//   - PostgreSQL
//   - NoSQL eg. MongoDB
//
// NOTE: this package is a adaptor to all implementations

type Store struct {
}

// New returns a new instance of InMemoryStore
func New() *InMemoryStore {
	return &InMemoryStore{}
}

type InMemoryStore struct {
	model  models.Snippet
	models []models.Snippet
}

// TODO: find a solution for not calling getLastID in multiple places
func (i *InMemoryStore) Add(text string, address string) (int, error) {
	i.models = append(i.models, models.Snippet{
		Id:      i.getLastID() + 1,
		Text:    text,
		Address: address,
	})

	return i.getLastID(), nil
}

func (i *InMemoryStore) Get(addr string) *models.Snippet {
	for _, model := range i.models {
		if model.Address == addr {
			return &model
		}
	}

	return nil
}

func (i *InMemoryStore) getLastID() int {
	if len(i.models) > 0 {
		return i.models[len(i.models)-1].Id
	}
	return 0
}
