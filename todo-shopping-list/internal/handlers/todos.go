package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	todoLists, err := h.db.GetTodoLists()
	if dbError(w, err) {
		return
	}
	shoppingLists, err := h.db.GetShoppingLists()
	if dbError(w, err) {
		return
	}

	data := map[string]interface{}{
		"ActiveNav":      "home",
		"TodoCount":      len(todoLists),
		"ShoppingCount":  len(shoppingLists),
		"RecentTodos":    limitLists(todoLists, 5),
		"RecentShopping": limitShoppingLists(shoppingLists, 5),
	}
	h.render(w, r, "templates/index.html", data)
}

func limitLists(lists []database.TodoList, n int) []database.TodoList {
	if len(lists) <= n {
		return lists
	}
	return lists[:n]
}

func limitShoppingLists(lists []database.ShoppingList, n int) []database.ShoppingList {
	if len(lists) <= n {
		return lists
	}
	return lists[:n]
}

func (h *Handler) handleTodoLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.db.GetTodoLists()
	if dbError(w, err) {
		return
	}
	categories, err := h.db.GetCategories()
	if dbError(w, err) {
		return
	}
	data := map[string]interface{}{
		"ActiveNav":  "todos",
		"Title":      "Todo Lists",
		"Lists":      lists,
		"Categories": categories,
	}
	h.render(w, r, "templates/todos.html", data)
}

func (h *Handler) handleCreateTodoList(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	list := &database.TodoList{Name: name}
	if catID := r.FormValue("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		if id > 0 {
			list.CategoryID = &id
		}
	}
	if dbError(w, h.db.CreateTodoList(list)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/todos/"+strconv.FormatInt(list.ID, 10)), http.StatusSeeOther)
}

func (h *Handler) handleTodoDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	list, err := h.db.GetTodoList(id)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	items, err := h.db.GetTodoItems(id)
	if dbError(w, err) {
		return
	}
	categories, err := h.db.GetCategories()
	if dbError(w, err) {
		return
	}
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
	data := map[string]interface{}{
		"ActiveNav":  "todos",
		"Title":      list.Name,
		"List":       list,
		"Items":      items,
		"Categories": categories,
		"Tags":       tags,
	}
	h.render(w, r, "templates/todo_detail.html", data)
}

func (h *Handler) handleUpdateTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	list := &database.TodoList{ID: id, Name: name}
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			list.CategoryID = &cid
		}
	}
	if dbError(w, h.db.UpdateTodoList(list)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/todos/"+strconv.FormatInt(id, 10)), http.StatusSeeOther)
}

func (h *Handler) handleDeleteTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if dbError(w, h.db.DeleteTodoList(id)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/todos"), http.StatusSeeOther)
}

func (h *Handler) handleCreateTodoItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	title, ok := requiredField(w, r, "title", "Title")
	if !ok {
		return
	}
	item := &database.TodoItem{
		ListID: listID,
		Title:  title,
	}
	if d := r.FormValue("deadline"); d != "" {
		item.Deadline = &d
	}
	if ds := r.FormValue("date_start"); ds != "" {
		item.DateStart = &ds
	}
	if de := r.FormValue("date_end"); de != "" {
		item.DateEnd = &de
	}
	if dbError(w, h.db.CreateTodoItemWithTags(item, formTagIDs(r))) {
		return
	}

	h.renderTodoItems(w, listID)
}

func (h *Handler) handleUpdateTodoItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	r.ParseForm()

	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	title, ok := requiredField(w, r, "title", "Title")
	if !ok {
		return
	}
	item.Title = title
	item.Description = strings.TrimSpace(r.FormValue("description"))
	if d := r.FormValue("deadline"); d != "" {
		item.Deadline = &d
	} else {
		item.Deadline = nil
	}
	if ds := r.FormValue("date_start"); ds != "" {
		item.DateStart = &ds
	} else {
		item.DateStart = nil
	}
	if de := r.FormValue("date_end"); de != "" {
		item.DateEnd = &de
	} else {
		item.DateEnd = nil
	}
	if dbError(w, h.db.UpdateTodoItemWithTags(item, formTagIDs(r))) {
		return
	}

	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	if dbError(w, h.db.DeleteTodoItem(itemID)) {
		return
	}
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleToggleTodoItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	if dbError(w, h.db.ToggleTodoItem(itemID)) {
		return
	}
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleSetTodoItemTags(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	r.ParseForm()
	if dbError(w, h.db.SetTodoItemTags(itemID, formTagIDs(r))) {
		return
	}
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleReorderTodoItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ItemIDs []int64 `json:"item_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if dbError(w, h.db.ReorderTodoItems(listID, body.ItemIDs)) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) renderTodoItems(w http.ResponseWriter, listID int64) {
	items, err := h.db.GetTodoItems(listID)
	if dbError(w, err) {
		return
	}
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
	h.renderPartial(w, "todo-items-list", map[string]interface{}{
		"Items": items,
		"Tags":  tags,
	})
}

func (h *Handler) handleGetTodoItemEdit(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
	var itemTagIDs []int64
	for _, t := range item.Tags {
		itemTagIDs = append(itemTagIDs, t.ID)
	}
	h.renderPartial(w, "todo-item-edit-form", map[string]interface{}{
		"Item":       item,
		"Tags":       tags,
		"ItemTagIDs": itemTagIDs,
	})
}
