package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/spaste/internal/paste"
)

func (s *server) createPaste(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		s.serverError(w, r, ErrEmptyCreatePasteBody, http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	var data paste.Module
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		s.serverError(w, r, fmt.Errorf("failed to decode body: %v", err), http.StatusInternalServerError)
		return
	}

	s.validator.IsBlank("Text", data.Text)
	s.validator.IsBlank("Lang", data.Lang)
	s.validator.IsEqual("LineCount", data.LineCount, 0)
	err = s.validator.Valid()
	if err != nil {
		s.serverError(w, r, err, http.StatusBadRequest)
		return
	}

	m := paste.Module{
		Address:   data.Address,
		Text:      data.Text,
		Lang:      data.Lang,
		LineCount: data.LineCount,
	}

	address, err := s.paste.Create(r, m)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	s.validator.IsBlank("Address", address)
	err = s.validator.Valid()
	if err != nil {
		s.serverError(w, r, err, http.StatusBadRequest)
		return
	}

	var addressData paste.Module
	addressData.Address = address

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(addressData)
	if err != nil {
		s.serverError(w, r, fmt.Errorf("failed to encode body: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *server) renderPaste(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "addr")

	if len(address) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data, err := s.paste.Get(r, address)
	if err == paste.ErrNoPasteFound {
		s.notFoundHandler(w, r)
		return
	} else if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	renderedPaste, err := s.paste.Render(data)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
		return
	}

	t := &paste.TemplateData{
		TextHighlighted: renderedPaste,
		Address:         fmt.Sprintf("%s/%s", r.Host, data.Address),
		LineCount:       data.LineCount,
		Lang:            data.Lang,
	}

	err = s.template.Render(w, "paste.page.tmpl", t)
	if err != nil {
		s.serverError(w, r, err, http.StatusInternalServerError)
	}
}
