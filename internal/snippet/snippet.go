package snippet

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/mortezadadgar/spaste/internal/models"
	"github.com/mortezadadgar/spaste/internal/validator"
)

// Create uses pointer for models as text can get really big.
type store interface {
	Create(snippet *models.Snippet) error
	Get(addr string) (*models.Snippet, error)
}

type Snippet struct {
	store store
	v     validator.Validator
}

// New returns a instance of SnippetStore.
func New(store store, v validator.Validator) Snippet {
	return Snippet{
		store: store,
		v:     v,
	}
}

// Create creates a new snippet in store.
func (s Snippet) Create(text string, lang string, lineCount int, adreess string) error {
	snippetModel := models.Snippet{
		Text:      text,
		Lang:      lang,
		LineCount: lineCount,
		Address:   adreess,
		TimeStamp: time.Now().Format(time.DateTime),
	}

	// sanity checks
	s.v.IsBlank("snippet.Text", snippetModel.Text)
	s.v.IsBlank("snippet.Address", snippetModel.Address)
	s.v.IsBlank("snippet.Lang", snippetModel.Lang)
	s.v.IsEqual("snippet.LineCount", snippetModel.LineCount, 0)
	if err := s.v.IsValid(); err != nil {
		return err
	}

	return s.store.Create(&snippetModel)
}

// Get gets snippet by its address.
func (s Snippet) Get(addr string) (*models.Snippet, error) {
	return s.store.Get(addr)
}

// MakeAddress makes a random address.
func (s Snippet) MakeAddress(length int64) (string, error) {
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
