package template

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"
)

type Template struct {
	dir         string
	hasLayout   bool
	templateMap map[string]*template.Template
}

// New returns a new instance of Template
//
// caller must have files named in *.page.tmpl format.
func New(dir string, hasLayout bool) (Template, error) {
	t := Template{
		dir:         dir,
		hasLayout:   hasLayout,
		templateMap: make(map[string]*template.Template),
	}

	if err := t.cacheTemplate(); err != nil {
		return Template{}, err
	}

	return t, nil
}

// Render executes template by its name.
func (t *Template) Render(w io.Writer, name string, data any) error {
	err := t.templateMap[name].ExecuteTemplate(w, name, data) // how
	if err != nil {
		return fmt.Errorf("failed to execute template name %s: %v", name, err)
	}

	return nil
}

func (t *Template) cacheTemplate() error {
	pages, err := filepath.Glob(filepath.Join(t.dir, "*.page.tmpl"))
	switch {
	case len(pages) < 1:
		return fmt.Errorf("no template page file found") // test
	case err != nil:
		return fmt.Errorf("failed to find template pages: %v", err) // test
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

	cache := t.templateMap
	for _, page := range pages {
		name := filepath.Base(page)
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(page) // how
		if err != nil {
			return fmt.Errorf("failed to parse template page %s: %v", page, err)
		}

		if t.hasLayout {
			tmpl, err = tmpl.ParseGlob(filepath.Join(t.dir, "*.layout.tmpl")) // how
			if err != nil {
				return fmt.Errorf("failed to parse layout templates: %v", err)
			}
		}

		cache[name] = tmpl
	}

	return nil
}
