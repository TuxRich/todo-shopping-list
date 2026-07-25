package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleSettings(w http.ResponseWriter, r *http.Request) {
	categories, err := h.db.GetCategories()
	if dbError(w, err) {
		return
	}
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
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
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	cat := &database.Category{
		Name:  name,
		Color: r.FormValue("color"),
		Icon:  strings.TrimSpace(r.FormValue("icon")),
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}
	if cat.Icon == "" {
		cat.Icon = "📁"
	}
	if dbError(w, h.db.CreateCategory(cat)) {
		return
	}
	h.renderCategories(w)
}

func (h *Handler) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	cat := &database.Category{
		ID:    id,
		Name:  name,
		Color: r.FormValue("color"),
		Icon:  strings.TrimSpace(r.FormValue("icon")),
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}
	if cat.Icon == "" {
		cat.Icon = "📁"
	}
	if dbError(w, h.db.UpdateCategory(cat)) {
		return
	}
	h.renderCategories(w)
}

func (h *Handler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if dbError(w, h.db.DeleteCategory(id)) {
		return
	}
	h.renderCategories(w)
}

func (h *Handler) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	tag := &database.Tag{
		Name:  name,
		Color: r.FormValue("color"),
	}
	if tag.Color == "" {
		tag.Color = "#8b5cf6"
	}
	if dbError(w, h.db.CreateTag(tag)) {
		return
	}
	h.renderTags(w)
}

func (h *Handler) handleUpdateTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	tag := &database.Tag{
		ID:    id,
		Name:  name,
		Color: r.FormValue("color"),
	}
	if tag.Color == "" {
		tag.Color = "#8b5cf6"
	}
	if dbError(w, h.db.UpdateTag(tag)) {
		return
	}
	h.renderTags(w)
}

func (h *Handler) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if dbError(w, h.db.DeleteTag(id)) {
		return
	}
	h.renderTags(w)
}

func (h *Handler) renderCategories(w http.ResponseWriter) {
	categories, err := h.db.GetCategories()
	if dbError(w, err) {
		return
	}
	h.renderPartial(w, "categories-list", categories)
}

func (h *Handler) renderTags(w http.ResponseWriter) {
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
	h.renderPartial(w, "tags-list", tags)
}
