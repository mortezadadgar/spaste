package template

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

type Template struct {
	dir         string
	hasLayout   bool
	templateMap map[string]*template.Template
}

// caller must have files named in *.page.tmpl format.
func New(dir string, hasLayout bool) (*Template, error) {
	r := &Template{
		dir:         dir,
		hasLayout:   hasLayout,
		templateMap: make(map[string]*template.Template),
	}

	err := r.cacheTemplate()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Template) Render(w http.ResponseWriter, name string, data any) error {
	err := r.templateMap[name].ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}

	return nil
}

func (r *Template) cacheTemplate() error {
	pages, err := filepath.Glob(filepath.Join(r.dir, "*.page.tmpl"))
	if len(pages) < 1 {
		return errors.New("no template page found")
	} else if err != nil {
		return err
	}

	cache := r.templateMap
	for _, page := range pages {
		name := filepath.Base(page)
		tmpl, err := template.New(name).ParseFiles(page)
		if err != nil {
			return err
		}

		if r.hasLayout {
			tmpl, err = tmpl.ParseGlob(filepath.Join(r.dir, "*.layout.tmpl"))
			if err != nil {
				return err
			}
		}

		cache[name] = tmpl
	}

	return nil
}
