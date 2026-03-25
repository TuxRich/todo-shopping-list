package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	todoLists, _ := h.db.GetTodoLists()
	shoppingLists, _ := h.db.GetShoppingLists()

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
	lists, _ := h.db.GetTodoLists()
	categories, _ := h.db.GetCategories()
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
	list := &database.TodoList{Name: r.FormValue("name")}
	if catID := r.FormValue("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		if id > 0 {
			list.CategoryID = &id
		}
	}
	h.db.CreateTodoList(list)
	http.Redirect(w, r, "/todos/"+strconv.FormatInt(list.ID, 10), http.StatusSeeOther)
}

func (h *Handler) handleTodoDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	list, err := h.db.GetTodoList(id)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	items, _ := h.db.GetTodoItems(id)
	categories, _ := h.db.GetCategories()
	tags, _ := h.db.GetTags()
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
	list := &database.TodoList{ID: id, Name: r.FormValue("name")}
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			list.CategoryID = &cid
		}
	}
	h.db.UpdateTodoList(list)
	http.Redirect(w, r, "/todos/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handler) handleDeleteTodoList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.DeleteTodoList(id)
	http.Redirect(w, r, "/todos", http.StatusSeeOther)
}

func (h *Handler) handleCreateTodoItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	item := &database.TodoItem{
		ListID: listID,
		Title:  r.FormValue("title"),
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
	h.db.CreateTodoItem(item)

	if tagIDs := r.Form["tag_ids"]; len(tagIDs) > 0 {
		var ids []int64
		for _, t := range tagIDs {
			tid, _ := strconv.ParseInt(t, 10, 64)
			ids = append(ids, tid)
		}
		h.db.SetTodoItemTags(item.ID, ids)
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

	item.Title = r.FormValue("title")
	item.Description = r.FormValue("description")
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
	h.db.UpdateTodoItem(item)

	var tagIDs []int64
	for _, t := range r.Form["tag_ids"] {
		tid, _ := strconv.ParseInt(t, 10, 64)
		tagIDs = append(tagIDs, tid)
	}
	h.db.SetTodoItemTags(item.ID, tagIDs)

	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	h.db.DeleteTodoItem(itemID)
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleToggleTodoItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetTodoItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	h.db.ToggleTodoItem(itemID)
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleSetTodoItemTags(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	r.ParseForm()
	var tagIDs []int64
	for _, t := range r.Form["tag_ids"] {
		tid, _ := strconv.ParseInt(t, 10, 64)
		tagIDs = append(tagIDs, tid)
	}
	h.db.SetTodoItemTags(itemID, tagIDs)
	item, _ := h.db.GetTodoItem(itemID)
	h.renderTodoItems(w, item.ListID)
}

func (h *Handler) handleReorderTodoItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ItemIDs []int64 `json:"item_ids"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	h.db.ReorderTodoItems(listID, body.ItemIDs)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) renderTodoItems(w http.ResponseWriter, listID int64) {
	items, _ := h.db.GetTodoItems(listID)
	tags, _ := h.db.GetTags()
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
	tags, _ := h.db.GetTags()
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
