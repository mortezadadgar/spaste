package snippets

import "github.com/mortezadadgar/spaste/models"

// TODO: make this package useful or remove it.
type SnippetStore interface {
	Add(text string, address string) (int, error)
	Get(addr string) *models.Snippet
}

type Snippet struct {
	store SnippetStore
}

// New returns a instance of SnippetStore.
func New(store SnippetStore) *Snippet {
	return &Snippet{
		store: store,
	}
}

func (s *Snippet) Add(text string, address string) (int, error) {
	return s.store.Add(text, address)
}

func (s *Snippet) Get(addr string) *models.Snippet {
	return s.store.Get(addr)
}

func (s *Snippet) Delete(int) error {
	return nil
}
