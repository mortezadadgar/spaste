package paste

// TODO: find a better name for this package.
import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/modules"
)

type store interface {
	Create(ctx context.Context, paste *modules.Paste) error
	Get(ctx context.Context, addr string) (*modules.Paste, error)
}

type Paste struct {
	store  store
	config config.Config
}

// New returns a instance of Paste.
func New(store store, config config.Config) *Paste {
	return &Paste{
		store:  store,
		config: config,
	}
}

// Create creates a new paste in store.
func (u Paste) Create(r *http.Request, text string, lang string, lineCount int) (string, error) {
	randomAddress, err := makeAddress(u.config.AddressLength, lang)
	if err != nil {
		return "", fmt.Errorf("failed to generate paste address: %v", err)
	}

	m := modules.Paste{
		Text:      text,
		Lang:      lang,
		LineCount: lineCount,
		Address:   randomAddress,
		TimeStamp: time.Now().Format(time.DateTime),
	}

	return m.Address, u.store.Create(r.Context(), &m)
}

// Get gets paste by its address.
func (u Paste) Get(r *http.Request, addr string) (*modules.Paste, error) {
	return u.store.Get(r.Context(), addr)
}

// Render render text in selected syntax highlighted language.
func (u Paste) Render(m *modules.Paste) (string, error) {
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
