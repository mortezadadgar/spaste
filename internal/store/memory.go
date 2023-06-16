package store

import "github.com/mortezadadgar/spaste/internal/modules"

type InMemoryStore struct {
	models []modules.Paste
}

// NewInMemoryStore returns a instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}

func (i *InMemoryStore) Create(p modules.Paste) error {
	i.models = append(i.models, modules.Paste{
		ID:        i.getLastID() + 1,
		Text:      p.Text,
		Lang:      p.Lang,
		LineCount: p.LineCount,
		Address:   p.Address,
	})

	return nil
}

func (i *InMemoryStore) Get(addr string) (modules.Paste, error) {
	for _, model := range i.models {
		if model.Address == addr {
			return modules.Paste{}, nil
		}
	}

	return modules.Paste{}, nil
}

func (i *InMemoryStore) getLastID() int {
	if len(i.models) > 0 {
		return i.models[len(i.models)-1].ID
	}
	return 0
}
