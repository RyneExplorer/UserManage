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
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return "?"
			}
			runes := []rune(s)
			return string(runes[0])
		},
		"add": func(a, b int) int { return a + b },
		// pageRange returns page numbers with -1 as ellipsis sentinel
		// e.g. page=5, total=10 → [1, -1, 4, 5, 6, -1, 10]
		"pageRange": func(page, total int) []int {
			if total < 1 {
				total = 1
			}
			set := map[int]bool{}
			// always include first, last, current and neighbours
			for _, p := range []int{1, 2, total - 1, total, page - 1, page, page + 1} {
				if p >= 1 && p <= total {
					set[p] = true
				}
			}
			// build sorted slice with -1 gaps
			result := []int{}
			prev := 0
			for i := 1; i <= total; i++ {
				if set[i] {
					if prev != 0 && i-prev > 1 {
						result = append(result, -1)
					}
					result = append(result, i)
					prev = i
				}
			}
			return result
		},
	}).ParseFiles(filepath.Join(r.Dir, name))
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.Execute(w, data)
}
