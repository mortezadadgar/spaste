package paste

// TODO: find a better name for this package.
import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/mortezadadgar/spaste/internal/config"
)

var ErrNoPasteFound = errors.New("no paste found")

type Module struct {
	ID        int    `json:"-"`
	Text      string `json:"text,omitempty"`
	Lang      string `json:"lang,omitempty"`
	LineCount int    `json:"linecount,omitempty"`
	Address   string `json:"address,omitempty"`
	TimeStamp string `json:"-"`
}

type TemplateData struct {
	Address         string
	TextHighlighted string
	LineCount       int
	Lang            string
	Message         string
	IncludeHome     bool
}

type store interface {
	Create(ctx context.Context, paste Module) error
	Get(ctx context.Context, addr string) (Module, error)
}

type Paste struct {
	store  store
	config config.Config
}

// New returns a instance of Paste.
func New(store store, config config.Config) Paste {
	return Paste{
		store:  store,
		config: config,
	}
}

// Create creates a new paste in store.
func (u Paste) Create(r *http.Request, m Module) (string, error) {
	address := m.Address
	var err error

	if len(m.Address) == 0 {
		address, err = makeAddress(u.config.AddressLength, m.Lang)
		if err != nil {
			return "", fmt.Errorf("failed to generate paste address: %v", err)
		}
	}

	m = Module{
		Text:      m.Text,
		Lang:      m.Lang,
		LineCount: m.LineCount,
		Address:   address,
		TimeStamp: time.Now().Format(time.DateTime),
	}

	return address, u.store.Create(r.Context(), m)
}

// Get gets paste by its address.
func (u Paste) Get(r *http.Request, addr string) (Module, error) {
	return u.store.Get(r.Context(), addr)
}

// Render render text in selected syntax highlighted language.
func (u Paste) Render(m Module) (string, error) {
	lexer := lexers.Get(m.Lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := html.New(html.WithClasses(true))

	iterator, err := lexer.Tokenise(nil, string(m.Text))
	if err != nil {
		return "", fmt.Errorf("failed to tokenise code: %v", err)
	}

	// TODO: why it's hardcoded.
	style := styles.Get("doom-one")
	if style == nil {
		style = styles.Fallback
	}

	buf := new(bytes.Buffer)
	err = formatter.Format(buf, style, iterator)
	if err != nil {
		return "", fmt.Errorf("failed to format code: %v", err)
	}

	return buf.String(), nil
}

func makeAddress(length int64, lang string) (string, error) {
	if length == 0 {
		return "", fmt.Errorf("generating empty address is not allowed")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	buffer := make([]byte, length)
	for i := range buffer {
		r, err := rand.Int(rand.Reader, big.NewInt(length))
		if err != nil {
			return "", fmt.Errorf("failed to generate random addresses: %v", err)
		}
		buffer[i] = charset[r.Int64()]
	}

	return fmt.Sprintf("%s.%s", buffer, lang), nil
}
