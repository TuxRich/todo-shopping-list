package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, map[string]string{"error": msg}, status)
}

// jsonPatch holds a decoded JSON object so updates can apply only the fields the
// client actually sent, leaving omitted columns untouched.
type jsonPatch struct {
	fields map[string]json.RawMessage
	err    error
}

func decodePatch(r *http.Request) (*jsonPatch, error) {
	var fields map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		return nil, err
	}
	return &jsonPatch{fields: fields}, nil
}

// apply writes the field into dst if it was present in the request body.
// A JSON null clears a pointer destination, so nullable columns stay settable.
func (p *jsonPatch) apply(key string, dst interface{}) {
	raw, ok := p.fields[key]
	if !ok || p.err != nil {
		return
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		p.err = fmt.Errorf("invalid value for %q", key)
	}
}

// --- Todo Lists ---

func (h *Handler) apiGetTodoLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.db.GetTodoLists()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lists == nil {
		lists = []database.TodoList{}
	}
	jsonResponse(w, lists, http.StatusOK)
}

func (h *Handler) apiCreateTodoList(w http.ResponseWriter, r *http.Request) {
	var list database.TodoList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	list.Name = strings.TrimSpace(list.Name)
	if list.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateTodoList(&list); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusCreated)
}

func (h *Handler) apiGetTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	list, err := h.db.GetTodoList(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

func (h *Handler) apiUpdateTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	list, err := h.db.GetTodoList(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("name", &list.Name)
	patch.apply("category_id", &list.CategoryID)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	list.Name = strings.TrimSpace(list.Name)
	if list.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.UpdateTodoList(list); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.db.GetTodoList(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, updated, http.StatusOK)
}

func (h *Handler) apiDeleteTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteTodoList(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) apiGetTodoItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	items, err := h.db.GetTodoItems(listID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []database.TodoItem{}
	}
	jsonResponse(w, items, http.StatusOK)
}

func (h *Handler) apiCreateTodoItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var item database.TodoItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item.ListID = listID
	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		jsonError(w, "Title is required", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateTodoItem(&item); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, item, http.StatusCreated)
}

func (h *Handler) apiGetTodoItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	item, err := h.db.GetTodoItem(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, item, http.StatusOK)
}

func (h *Handler) apiUpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.db.GetTodoItem(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("title", &item.Title)
	patch.apply("description", &item.Description)
	patch.apply("completed", &item.Completed)
	patch.apply("deadline", &item.Deadline)
	patch.apply("date_start", &item.DateStart)
	patch.apply("date_end", &item.DateEnd)
	patch.apply("sort_order", &item.SortOrder)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		jsonError(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := h.db.UpdateTodoItem(item); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.db.GetTodoItem(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, updated, http.StatusOK)
}

func (h *Handler) apiDeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteTodoItem(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Shopping Lists ---

func (h *Handler) apiGetShoppingLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.db.GetShoppingLists()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lists == nil {
		lists = []database.ShoppingList{}
	}
	jsonResponse(w, lists, http.StatusOK)
}

func (h *Handler) apiCreateShoppingList(w http.ResponseWriter, r *http.Request) {
	var list database.ShoppingList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	list.Name = strings.TrimSpace(list.Name)
	if list.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateShoppingList(&list); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusCreated)
}

func (h *Handler) apiGetShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	list, err := h.db.GetShoppingList(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

func (h *Handler) apiUpdateShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	list, err := h.db.GetShoppingList(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("name", &list.Name)
	patch.apply("category_id", &list.CategoryID)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	list.Name = strings.TrimSpace(list.Name)
	if list.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.UpdateShoppingList(list); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.db.GetShoppingList(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, updated, http.StatusOK)
}

func (h *Handler) apiDeleteShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteShoppingList(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) apiGetShoppingItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	sort := r.URL.Query().Get("sort")
	items, err := h.db.GetShoppingItems(listID, sort)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []database.ShoppingItem{}
	}
	jsonResponse(w, items, http.StatusOK)
}

func (h *Handler) apiCreateShoppingItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var item database.ShoppingItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item.ListID = listID
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if item.Quantity < 1 {
		item.Quantity = 1
	}

	if existing, err := h.db.FindShoppingItemByName(listID, item.Name); err == nil {
		h.db.ReactivateShoppingItem(existing.ID, item.Quantity, item.Unit)
		reactivated, _ := h.db.GetShoppingItem(existing.ID)
		jsonResponse(w, reactivated, http.StatusOK)
		return
	}

	if err := h.db.CreateShoppingItem(&item); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, item, http.StatusCreated)
}

func (h *Handler) apiGetShoppingItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	item, err := h.db.GetShoppingItem(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, item, http.StatusOK)
}

func (h *Handler) apiUpdateShoppingItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := h.db.GetShoppingItem(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("name", &item.Name)
	patch.apply("quantity", &item.Quantity)
	patch.apply("unit", &item.Unit)
	patch.apply("purchased", &item.Purchased)
	patch.apply("category_id", &item.CategoryID)
	patch.apply("sort_order", &item.SortOrder)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if item.Quantity < 1 {
		item.Quantity = 1
	}

	if err := h.db.UpdateShoppingItem(item); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	updated, err := h.db.GetShoppingItem(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, updated, http.StatusOK)
}

func (h *Handler) apiDeleteShoppingItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteShoppingItem(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) apiSearchShoppingHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results, err := h.db.SearchShoppingHistory(q)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []database.ShoppingHistory{}
	}
	jsonResponse(w, results, http.StatusOK)
}

// --- Categories ---

func (h *Handler) apiGetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.db.GetCategories()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cats == nil {
		cats = []database.Category{}
	}
	jsonResponse(w, cats, http.StatusOK)
}

func (h *Handler) apiCreateCategory(w http.ResponseWriter, r *http.Request) {
	var cat database.Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cat.Name = strings.TrimSpace(cat.Name)
	if cat.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if cat.Color == "" {
		cat.Color = "#6366f1"
	}
	if cat.Icon == "" {
		cat.Icon = "📁"
	}
	if err := h.db.CreateCategory(&cat); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, cat, http.StatusCreated)
}

func (h *Handler) apiGetCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	cat, err := h.db.GetCategory(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, cat, http.StatusOK)
}

func (h *Handler) apiUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cat, err := h.db.GetCategory(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("name", &cat.Name)
	patch.apply("color", &cat.Color)
	patch.apply("icon", &cat.Icon)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	cat.Name = strings.TrimSpace(cat.Name)
	if cat.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.UpdateCategory(cat); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, cat, http.StatusOK)
}

func (h *Handler) apiDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteCategory(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Tags ---

func (h *Handler) apiGetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.db.GetTags()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []database.Tag{}
	}
	jsonResponse(w, tags, http.StatusOK)
}

func (h *Handler) apiCreateTag(w http.ResponseWriter, r *http.Request) {
	var tag database.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if tag.Color == "" {
		tag.Color = "#8b5cf6"
	}
	if err := h.db.CreateTag(&tag); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tag, http.StatusCreated)
}

func (h *Handler) apiGetTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	tag, err := h.db.GetTag(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, tag, http.StatusOK)
}

func (h *Handler) apiUpdateTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	patch, err := decodePatch(r)
	if err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	tag, err := h.db.GetTag(id)
	if err != nil {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}
	patch.apply("name", &tag.Name)
	patch.apply("color", &tag.Color)
	if patch.err != nil {
		jsonError(w, patch.err.Error(), http.StatusBadRequest)
		return
	}
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		jsonError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if err := h.db.UpdateTag(tag); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tag, http.StatusOK)
}

func (h *Handler) apiDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.db.DeleteTag(id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
