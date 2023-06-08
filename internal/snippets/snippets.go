package snippets

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/mortezadadgar/spaste/internal/models"
)

type SnippetStore interface {
	Add(snippet *models.Snippet) error
	Get(addr string) (*models.Snippet, error)
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

func (s *Snippet) Add(text string, lang string, lineCount int, adreess string) error {
	snippetModel := models.Snippet{
		Text:      text,
		Lang:      lang,
		LineCount: lineCount,
		Address:   adreess,
		TimeStamp: time.Now().Format(time.DateTime),
	}

	return s.store.Add(&snippetModel)
}

func (s *Snippet) Get(addr string) (*models.Snippet, error) {
	return s.store.Get(addr)
}

func (s *Snippet) Delete(int) error {
	// TODO: make snippets disposable
	return nil
}

func (s *Snippet) MakeAddress(length int64) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buffer := make([]byte, length)
	for i := range buffer {
		r, err := rand.Int(rand.Reader, big.NewInt(length))
		if err != nil {
			return "", fmt.Errorf("failed to generate random addresses: %v", err)
		}
		buffer[i] = letters[r.Int64()]
	}

	return string(buffer), nil
}
