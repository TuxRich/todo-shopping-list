package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"todo-app/internal/database"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleShoppingLists(w http.ResponseWriter, r *http.Request) {
	lists, _ := h.db.GetShoppingLists()
	categories, _ := h.db.GetCategories()
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
	list := &database.ShoppingList{Name: r.FormValue("name")}
	if catID := r.FormValue("category_id"); catID != "" {
		id, _ := strconv.ParseInt(catID, 10, 64)
		if id > 0 {
			list.CategoryID = &id
		}
	}
	h.db.CreateShoppingList(list)
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
	items, _ := h.db.GetShoppingItems(id, sort)
	categories, _ := h.db.GetCategories()
	tags, _ := h.db.GetTags()
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
	list := &database.ShoppingList{ID: id, Name: r.FormValue("name")}
	if catID := r.FormValue("category_id"); catID != "" {
		cid, _ := strconv.ParseInt(catID, 10, 64)
		if cid > 0 {
			list.CategoryID = &cid
		}
	}
	h.db.UpdateShoppingList(list)
	http.Redirect(w, r, redirectURL(r, "/shopping/"+strconv.FormatInt(id, 10)), http.StatusSeeOther)
}

func (h *Handler) handleDeleteShoppingList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.DeleteShoppingList(id)
	http.Redirect(w, r, redirectURL(r, "/shopping"), http.StatusSeeOther)
}

func (h *Handler) handleCreateShoppingItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	r.ParseForm()
	name := r.FormValue("name")
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}
	unit := r.FormValue("unit")

	if existing, err := h.db.FindShoppingItemByName(listID, name); err == nil {
		h.db.ReactivateShoppingItem(existing.ID, qty, unit)
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
	h.db.CreateShoppingItem(item)

	if tagIDs := r.Form["tag_ids"]; len(tagIDs) > 0 {
		var ids []int64
		for _, t := range tagIDs {
			tid, _ := strconv.ParseInt(t, 10, 64)
			ids = append(ids, tid)
		}
		h.db.SetShoppingItemTags(item.ID, ids)
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
	categories, _ := h.db.GetCategories()
	tags, _ := h.db.GetTags()
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
	item.Name = r.FormValue("name")
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}
	item.Quantity = qty
	item.Unit = r.FormValue("unit")
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
	h.db.UpdateShoppingItem(item)

	var tagIDs []int64
	for _, t := range r.Form["tag_ids"] {
		tid, _ := strconv.ParseInt(t, 10, 64)
		tagIDs = append(tagIDs, tid)
	}
	h.db.SetShoppingItemTags(item.ID, tagIDs)

	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleDeleteShoppingItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	h.db.DeleteShoppingItem(itemID)
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleToggleShoppingItem(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	item, err := h.db.GetShoppingItem(itemID)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}
	h.db.ToggleShoppingItem(itemID)
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleSetShoppingItemTags(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	r.ParseForm()
	var tagIDs []int64
	for _, t := range r.Form["tag_ids"] {
		tid, _ := strconv.ParseInt(t, 10, 64)
		tagIDs = append(tagIDs, tid)
	}
	h.db.SetShoppingItemTags(itemID, tagIDs)
	item, _ := h.db.GetShoppingItem(itemID)
	h.renderShoppingItems(w, r, item.ListID)
}

func (h *Handler) handleReorderShoppingItems(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ItemIDs []int64 `json:"item_ids"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	h.db.ReorderShoppingItems(listID, body.ItemIDs)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleSearchShoppingHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("name")
	if q == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	results, _ := h.db.SearchShoppingHistory(q)
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
	items, _ := h.db.GetShoppingItems(listID, sort)
	tags, _ := h.db.GetTags()
	h.renderPartial(w, "shopping-items-list", map[string]interface{}{
		"Items": items,
		"Tags":  tags,
		"Sort":  sort,
	})
}
