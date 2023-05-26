package snippets

import "github.com/mortezadadgar/spaste/models"

// TODO: make this package useful or remove it
type SnippetStore interface {
	Add(text string, address string) (int, error)
	Get(addr string) *models.Snippet
}

type Snippet struct {
	store SnippetStore
}

func NewSnippets(store SnippetStore) (*Snippet, error) {
	return &Snippet{
		store: store,
	}, nil
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
