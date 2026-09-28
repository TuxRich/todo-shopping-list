package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleShoppingLists(w http.ResponseWriter, r *http.Request) {
	lists, err := h.db.GetShoppingLists()
	if dbError(w, err) {
		return
	}
	categories, err := h.db.GetCategories()
	if dbError(w, err) {
		return
	}
	data := map[string]interface{}{
		"ActiveNav":  "shopping",
		"Title":      "Shopping Lists",
		"Lists":      lists,
		"Categories": categories,
	}
	h.render(w, r, "templates/shopping.html", data)
}

func (h *Handler) handleCreateShoppingList(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	list := &database.ShoppingList{Name: name}
	if catID := r.FormValue("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		if id > 0 {
			list.CategoryID = &id
		}
	}
	if dbError(w, h.db.CreateShoppingList(list)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/shopping/"+strconv.FormatInt(list.ID, 10)), http.StatusSeeOther)
}

func (h *Handler) handleShoppingDetail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	list, err := h.db.GetShoppingList(id)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	sort := r.URL.Query().Get("sort")
	items, err := h.db.GetShoppingItems(id, sort)
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
		"ActiveNav":  "shopping",
		"Title":      list.Name,
		"List":       list,
		"Items":      items,
		"Categories": categories,
		"Tags":       tags,
		"Sort":       sort,
	}
	h.render(w, r, "templates/shopping_detail.html", data)
}

func (h *Handler) handleUpdateShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	list := &database.ShoppingList{ID: id, Name: name}
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			list.CategoryID = &cid
		}
	}
	if dbError(w, h.db.UpdateShoppingList(list)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/shopping/"+strconv.FormatInt(id, 10)), http.StatusSeeOther)
}

func (h *Handler) handleDeleteShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if dbError(w, h.db.DeleteShoppingList(id)) {
		return
	}
	http.Redirect(w, r, redirectURL(r, "/shopping"), http.StatusSeeOther)
}

func (h *Handler) handleCreateShoppingItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}
	unit := strings.TrimSpace(r.FormValue("unit"))

	existing, err := h.db.FindShoppingItemByName(listID, name)
	if dbError(w, err) {
		return
	}
	if existing != nil {
		if dbError(w, h.db.ReactivateShoppingItem(existing.ID, qty, unit)) {
			return
		}
		h.renderShoppingItems(w, r, listID)
		return
	}

	item := &database.ShoppingItem{
		ListID:   listID,
		Name:     name,
		Quantity: qty,
		Unit:     unit,
	}
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			item.CategoryID = &cid
		}
	}
	if dbError(w, h.db.CreateShoppingItemWithTags(item, formTagIDs(r))) {
		return
	}

	h.renderShoppingItems(w, r, listID)
}

func (h *Handler) handleGetShoppingItemEdit(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
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
	var itemTagIDs []int64
	for _, t := range item.Tags {
		itemTagIDs = append(itemTagIDs, t.ID)
	}
	h.renderPartial(w, "shopping-item-edit-form", map[string]interface{}{
		"Item":       item,
		"Categories": categories,
		"Tags":       tags,
		"ItemTagIDs": itemTagIDs,
	})
}

func (h *Handler) handleUpdateShoppingItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)

	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	r.ParseForm()
	name, ok := requiredField(w, r, "name", "Name")
	if !ok {
		return
	}
	item.Name = name
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}
	item.Quantity = qty
	item.Unit = strings.TrimSpace(r.FormValue("unit"))
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			item.CategoryID = &cid
		} else {
			item.CategoryID = nil
		}
	} else {
		item.CategoryID = nil
	}
	if dbError(w, h.db.UpdateShoppingItemWithTags(item, formTagIDs(r))) {
		return
	}

	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleDeleteShoppingItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	if dbError(w, h.db.DeleteShoppingItem(itemID)) {
		return
	}
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleToggleShoppingItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	if dbError(w, h.db.ToggleShoppingItem(itemID)) {
		return
	}
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleSetShoppingItemTags(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	r.ParseForm()
	if dbError(w, h.db.SetShoppingItemTags(itemID, formTagIDs(r))) {
		return
	}
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleReorderShoppingItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ItemIDs []int64 `json:"item_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if dbError(w, h.db.ReorderShoppingItems(listID, body.ItemIDs)) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleSearchShoppingHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("name")
	if q == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	results, err := h.db.SearchShoppingHistory(q)
	if dbError(w, err) {
		return
	}
	h.renderPartial(w, "shopping-history-results", results)
}

func (h *Handler) renderShoppingItems(w http.ResponseWriter, r *http.Request, listID int64) {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		if ref := r.Header.Get("Hx-Current-Url"); ref != "" {
			if u, err := url.Parse(ref); err == nil {
				sort = u.Query().Get("sort")
			}
		}
	}
	items, err := h.db.GetShoppingItems(listID, sort)
	if dbError(w, err) {
		return
	}
	tags, err := h.db.GetTags()
	if dbError(w, err) {
		return
	}
	h.renderPartial(w, "shopping-items-list", map[string]interface{}{
		"Items": items,
		"Tags":  tags,
		"Sort":  sort,
	})
}
