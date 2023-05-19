package snippets

import "github.com/mortezadadgar/spaste/models"

// TODO: make to package useful or remove it
type SnippetStore interface {
	Add(text string, address string) (int, error)
	Get(id int) (*models.Snippet, error)
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

func (s *Snippet) Get(id int) (*models.Snippet, error) {
	return s.store.Get(id)
}

func (s *Snippet) Delete(int) error {
	return nil
}
