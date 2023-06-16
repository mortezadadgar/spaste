package server

// TODO: mock logs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/modules"
)

type mockPaste struct {
	CreateFunc func() (string, error)
	RenderFunc func() (string, error)
	GetFunc    func() (*modules.Paste, error)
}

func (s *mockPaste) Get(string) (*modules.Paste, error) {
	return s.GetFunc()
}

func (s *mockPaste) Create(string, string, int) (string, error) {
	return s.CreateFunc()
}

func (s *mockPaste) Render(*modules.Paste) (string, error) {
	return s.RenderFunc()
}

type stubTemplate struct {
	dummyError error
}

func (s *stubTemplate) Render(io.Writer, string, any) error {
	return s.dummyError
}

type stubValidator struct {
	dummyError error
}

func (s stubValidator) IsBlank(string, string) {
}

func (s stubValidator) IsEqual(string, int, int) {
}

func (s stubValidator) Valid() error {
	return s.dummyError
}

var errDummy = errors.New("bad errors")

const expectedAddress = "acajgcifig.text"

func TestCreatePaste(t *testing.T) {
	config, _ := config.New()
	template := stubTemplate{}

	pasteData := modules.Paste{
		Text:      "Hello world!",
		Lang:      "text",
		LineCount: 1,
	}

	t.Run("returns 200 on valid json data", func(t *testing.T) {
		validator := stubValidator{}
		paste := mockPaste{
			CreateFunc: func() (string, error) {
				return expectedAddress, nil
			},
		}

		server := New(config, &template, &paste, validator)

		b := marshalPaste(t, pasteData)
		request := newPasteRequest(bytes.NewReader(b))
		response := httptest.NewRecorder()

		server.createPaste(response, request)

		var addressData modules.Paste
		unmarshalPaste(t, response.Body, &addressData)

		assertPaste(t, addressData.Address, expectedAddress)
		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 500 on invalid json data", func(t *testing.T) {
		paste := mockPaste{}
		validator := stubValidator{}

		server := New(config, &template, &paste, validator)

		request := newPasteRequest(strings.NewReader("error"))
		response := httptest.NewRecorder()

		server.createPaste(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})

	t.Run("returns 404 on empty body", func(t *testing.T) {
		paste := mockPaste{}
		validator := stubValidator{}

		server := New(config, &template, &paste, validator)

		request := newPasteRequest(nil)
		response := httptest.NewRecorder()

		server.createPaste(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("returns 400 on validations", func(t *testing.T) {
		validator := stubValidator{
			dummyError: errDummy,
		}
		paste := mockPaste{
			CreateFunc: func() (string, error) {
				return expectedAddress, nil
			},
		}

		server := New(config, &template, &paste, validator)

		b := marshalPaste(t, pasteData)
		request := newPasteRequest(bytes.NewReader(b))
		response := httptest.NewRecorder()

		server.createPaste(response, request)

		assertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns 500 on creating paste", func(t *testing.T) {
		validator := stubValidator{}
		paste := mockPaste{
			CreateFunc: func() (string, error) {
				return expectedAddress, errDummy
			},
		}

		server := New(config, &template, &paste, validator)

		b := marshalPaste(t, pasteData)
		request := newPasteRequest(bytes.NewReader(b))
		response := httptest.NewRecorder()

		server.createPaste(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func TestRenderPaste(t *testing.T) {
	config, _ := config.New()
	template := stubTemplate{}

	t.Run("returns 200 on valid address", func(t *testing.T) {
		paste := mockPaste{
			GetFunc: func() (*modules.Paste, error) {
				return &modules.Paste{Address: expectedAddress}, nil
			},
			RenderFunc: func() (string, error) {
				return "", nil
			},
		}

		server := New(config, &template, &paste, nil)

		request := newPasteAddrRequest(expectedAddress)
		response := httptest.NewRecorder()

		server.renderPaste(response, request)

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 404 on empty address", func(t *testing.T) {
		paste := mockPaste{}

		server := New(config, &template, &paste, nil)

		request := newPasteAddrRequest("")
		response := httptest.NewRecorder()

		server.renderPaste(response, request)

		assertStatus(t, response.Code, http.StatusSeeOther)
	})

	t.Run("returns 404 on non-exist address", func(t *testing.T) {
		paste := mockPaste{
			GetFunc: func() (*modules.Paste, error) {
				return nil, nil
			},
		}

		server := New(config, &template, &paste, nil)

		request := newPasteAddrRequest(expectedAddress)
		response := httptest.NewRecorder()

		server.renderPaste(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("returns 500 on get paste errors", func(t *testing.T) {
		paste := mockPaste{
			GetFunc: func() (*modules.Paste, error) {
				return &modules.Paste{Address: expectedAddress}, errDummy
			},
		}

		server := New(config, &template, &paste, nil)

		request := newPasteAddrRequest(expectedAddress)
		response := httptest.NewRecorder()

		server.renderPaste(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})

	t.Run("returns 500 on render error", func(t *testing.T) {
		paste := mockPaste{
			GetFunc: func() (*modules.Paste, error) {
				return &modules.Paste{Address: expectedAddress}, nil
			},
			RenderFunc: func() (string, error) {
				return "", errDummy
			},
		}

		server := New(config, &template, &paste, nil)

		request := newPasteAddrRequest(expectedAddress)
		response := httptest.NewRecorder()

		server.renderPaste(response, request)

		assertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func assertPaste(t *testing.T, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertStatus(t *testing.T, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("wrong status code: got %d, want %d", got, want)
	}
}

func marshalPaste(t *testing.T, module modules.Paste) []byte {
	t.Helper()
	b, err := json.Marshal(module)
	if err != nil {
		t.Fatal(err)
	}

	return b

}

func unmarshalPaste(t *testing.T, body io.Reader, module *modules.Paste) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&module)
	if err != nil {
		t.Fatal(err)
	}
}

func newPasteAddrRequest(address string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("addr", address)
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	return request
}

func newPasteRequest(data io.Reader) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/paste", data)
	return request
}
