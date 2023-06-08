package template

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	ErrTemplateFileNotFound = errors.New("no template page file found")
)

type Template struct {
	dir         string
	hasLayout   bool
	templateMap map[string]*template.Template
}

var Data struct {
	Address         string
	TextHighlighted template.HTML
	LineCount       int
}

func ToHTML(s string) template.HTML {
	return template.HTML(s)
}

// New returns a new instance of Template
//
// caller must have files named in *.page.tmpl format.
func New(dir string, hasLayout bool) (*Template, error) {
	r := &Template{
		dir:         dir,
		hasLayout:   hasLayout,
		templateMap: make(map[string]*template.Template),
	}

	if err := r.cacheTemplate(); err != nil {
		return nil, err
	}

	return r, nil
}

// Rende executes template by its name.
func (r *Template) Render(w http.ResponseWriter, name string, data any) error {
	err := r.templateMap[name].ExecuteTemplate(w, name, data)
	if err != nil {
		return fmt.Errorf("failed to execute template name %s: %v", name, err)
	}

	return nil
}

func (r *Template) cacheTemplate() error {
	pages, err := filepath.Glob(filepath.Join(r.dir, "*.page.tmpl"))
	switch {
	case len(pages) < 1:
		return ErrTemplateFileNotFound
	case err != nil:
		return fmt.Errorf("failed to find template pages: %v", err)
	}

	funcMap := template.FuncMap{
		// increment values
		"inc": func(i int) int {
			return i + 1
		},
		// make integer iterable
		"makeSlice": func(i int) []int {
			return make([]int, i)
		},
	}

	cache := r.templateMap
	for _, page := range pages {
		name := filepath.Base(page)
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(page)
		if err != nil {
			return fmt.Errorf("failed to parse template page %s: %v", page, err)
		}

		if r.hasLayout {
			tmpl, err = tmpl.ParseGlob(filepath.Join(r.dir, "*.layout.tmpl"))
			if err != nil {
				return fmt.Errorf("failed to parse layout templates: %v", err)
			}
		}

		cache[name] = tmpl
	}

	return nil
}
