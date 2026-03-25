package handlers

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const ingressPathKey contextKey = "ingressPath"

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Handler struct {
	db    *database.DB
	pages map[string]*template.Template
	parts *template.Template
}

func NewRouter(db *database.DB) http.Handler {
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"contains": func(slice []int64, val int64) bool {
			for _, v := range slice {
				if v == val {
					return true
				}
			}
			return false
		},
		"join": strings.Join,
		"pct": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return (a * 100) / b
		},
		"derefStr": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"derefInt64": func(i *int64) int64 {
			if i == nil {
				return 0
			}
			return *i
		},
	}

	base := template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/base.html",
		"templates/partials/*.html",
	))

	pageFiles := []string{
		"templates/index.html",
		"templates/todos.html",
		"templates/todo_detail.html",
		"templates/shopping.html",
		"templates/shopping_detail.html",
		"templates/settings.html",
	}

	pages := make(map[string]*template.Template, len(pageFiles))
	for _, pf := range pageFiles {
		pt := template.Must(template.Must(base.Clone()).ParseFS(templateFS, pf))
		pages[pf] = pt
	}

	h := &Handler{db: db, pages: pages, parts: base}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(ingressMiddleware)

	staticContent, _ := fs.Sub(staticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	r.Get("/", h.handleIndex)

	r.Route("/todos", func(r chi.Router) {
		r.Get("/", h.handleTodoLists)
		r.Post("/", h.handleCreateTodoList)
		r.Get("/{id}", h.handleTodoDetail)
		r.Put("/{id}", h.handleUpdateTodoList)
		r.Delete("/{id}", h.handleDeleteTodoList)
		r.Post("/{id}/items", h.handleCreateTodoItem)
		r.Get("/items/{itemID}", h.handleGetTodoItemEdit)
		r.Put("/items/{itemID}", h.handleUpdateTodoItem)
		r.Delete("/items/{itemID}", h.handleDeleteTodoItem)
		r.Post("/items/{itemID}/toggle", h.handleToggleTodoItem)
		r.Put("/items/{itemID}/tags", h.handleSetTodoItemTags)
		r.Post("/{id}/reorder", h.handleReorderTodoItems)
	})

	r.Route("/shopping", func(r chi.Router) {
		r.Get("/", h.handleShoppingLists)
		r.Post("/", h.handleCreateShoppingList)
		r.Get("/{id}", h.handleShoppingDetail)
		r.Put("/{id}", h.handleUpdateShoppingList)
		r.Delete("/{id}", h.handleDeleteShoppingList)
		r.Post("/{id}/items", h.handleCreateShoppingItem)
		r.Get("/items/{itemID}", h.handleGetShoppingItemEdit)
		r.Put("/items/{itemID}", h.handleUpdateShoppingItem)
		r.Delete("/items/{itemID}", h.handleDeleteShoppingItem)
		r.Post("/items/{itemID}/toggle", h.handleToggleShoppingItem)
		r.Put("/items/{itemID}/tags", h.handleSetShoppingItemTags)
		r.Post("/{id}/reorder", h.handleReorderShoppingItems)
		r.Get("/history/search", h.handleSearchShoppingHistory)
	})

	r.Route("/settings", func(r chi.Router) {
		r.Get("/", h.handleSettings)
		r.Post("/categories", h.handleCreateCategory)
		r.Put("/categories/{id}", h.handleUpdateCategory)
		r.Delete("/categories/{id}", h.handleDeleteCategory)
		r.Post("/tags", h.handleCreateTag)
		r.Put("/tags/{id}", h.handleUpdateTag)
		r.Delete("/tags/{id}", h.handleDeleteTag)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/todo-lists", func(r chi.Router) {
			r.Get("/", h.apiGetTodoLists)
			r.Post("/", h.apiCreateTodoList)
			r.Get("/{id}", h.apiGetTodoList)
			r.Put("/{id}", h.apiUpdateTodoList)
			r.Delete("/{id}", h.apiDeleteTodoList)
			r.Get("/{id}/items", h.apiGetTodoItems)
			r.Post("/{id}/items", h.apiCreateTodoItem)
		})
		r.Route("/todo-items", func(r chi.Router) {
			r.Get("/{id}", h.apiGetTodoItem)
			r.Put("/{id}", h.apiUpdateTodoItem)
			r.Delete("/{id}", h.apiDeleteTodoItem)
		})
		r.Route("/shopping-lists", func(r chi.Router) {
			r.Get("/", h.apiGetShoppingLists)
			r.Post("/", h.apiCreateShoppingList)
			r.Get("/{id}", h.apiGetShoppingList)
			r.Put("/{id}", h.apiUpdateShoppingList)
			r.Delete("/{id}", h.apiDeleteShoppingList)
			r.Get("/{id}/items", h.apiGetShoppingItems)
			r.Post("/{id}/items", h.apiCreateShoppingItem)
		})
		r.Route("/shopping-items", func(r chi.Router) {
			r.Get("/{id}", h.apiGetShoppingItem)
			r.Put("/{id}", h.apiUpdateShoppingItem)
			r.Delete("/{id}", h.apiDeleteShoppingItem)
		})
		r.Get("/shopping-history", h.apiSearchShoppingHistory)
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", h.apiGetCategories)
			r.Post("/", h.apiCreateCategory)
			r.Get("/{id}", h.apiGetCategory)
			r.Put("/{id}", h.apiUpdateCategory)
			r.Delete("/{id}", h.apiDeleteCategory)
		})
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", h.apiGetTags)
			r.Post("/", h.apiCreateTag)
			r.Get("/{id}", h.apiGetTag)
			r.Put("/{id}", h.apiUpdateTag)
			r.Delete("/{id}", h.apiDeleteTag)
		})
	})

	return r
}

func ingressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ingressPath := r.Header.Get("X-Ingress-Path")
		if ingressPath != "" {
			ctx := context.WithValue(r.Context(), ingressPathKey, ingressPath)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func getIngressPath(r *http.Request) string {
	if v, ok := r.Context().Value(ingressPathKey).(string); ok {
		return v
	}
	return ""
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, page string, data map[string]interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	basePath := getIngressPath(r)
	if basePath != "" && !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	data["BasePath"] = basePath

	tmpl, ok := h.pages[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) renderPartial(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.parts.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
