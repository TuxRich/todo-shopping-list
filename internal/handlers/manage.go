package handlers

import (
	"net/http"
	"strconv"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	categories, _ := h.db.GetCategories()
	tags, _ := h.db.GetTags()
	data := map[string]interface{}{
		"ActiveNav":  "settings",
		"Title":      "Settings",
		"Categories": categories,
		"Tags":       tags,
	}
	h.render(w, r, "templates/settings.html", data)
}

func (h *Handler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cat := &database.Category{
		Name:  r.FormValue("name"),
		Color: r.FormValue("color"),
		Icon:  r.FormValue("icon"),
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}
	if cat.Icon == "" {
		cat.Icon = "📁"
	}
	h.db.CreateCategory(cat)
	categories, _ := h.db.GetCategories()
	h.renderPartial(w, "categories-list", categories)
}

func (h *Handler) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	cat := &database.Category{
		ID:    id,
		Name:  r.FormValue("name"),
		Color: r.FormValue("color"),
		Icon:  r.FormValue("icon"),
	}
	h.db.UpdateCategory(cat)
	categories, _ := h.db.GetCategories()
	h.renderPartial(w, "categories-list", categories)
}

func (h *Handler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.DeleteCategory(id)
	categories, _ := h.db.GetCategories()
	h.renderPartial(w, "categories-list", categories)
}

func (h *Handler) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tag := &database.Tag{
		Name:  r.FormValue("name"),
		Color: r.FormValue("color"),
	}
	if tag.Color == "" {
		tag.Color = "#8b5cf6"
	}
	h.db.CreateTag(tag)
	tags, _ := h.db.GetTags()
	h.renderPartial(w, "tags-list", tags)
}

func (h *Handler) handleUpdateTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	tag := &database.Tag{
		ID:    id,
		Name:  r.FormValue("name"),
		Color: r.FormValue("color"),
	}
	h.db.UpdateTag(tag)
	tags, _ := h.db.GetTags()
	h.renderPartial(w, "tags-list", tags)
}

func (h *Handler) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.DeleteTag(id)
	tags, _ := h.db.GetTags()
	h.renderPartial(w, "tags-list", tags)
}
