package view

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type Renderer struct {
	Dir string
}

func NewRenderer(dir string) *Renderer {
	return &Renderer{Dir: dir}
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) error {
	t, err := template.New(name).Funcs(template.FuncMap{
		"eq": func(a any, b any) bool { return a == b },
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02 15:04")
		},
	}).ParseFiles(filepath.Join(r.Dir, name))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.Execute(w, data)
}
