package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (db *DB) GetShoppingLists() ([]ShoppingList, error) {
	rows, err := db.conn.Query(`
		SELECT sl.id, sl.name, sl.category_id, sl.created_at, sl.updated_at,
			c.id, c.name, c.color, c.icon,
			COUNT(si.id) AS item_count,
			COUNT(CASE WHEN si.purchased = 1 THEN 1 END) AS done_count
		FROM shopping_lists sl
		LEFT JOIN categories c ON sl.category_id = c.id
		LEFT JOIN shopping_items si ON si.list_id = sl.id
		GROUP BY sl.id
		ORDER BY sl.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query shopping lists: %w", err)
	}
	defer rows.Close()

	var lists []ShoppingList
	for rows.Next() {
		var l ShoppingList
		var catIDInt sql.NullInt64
		var catName, catColor, catIcon sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &l.CategoryID, &l.CreatedAt, &l.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon,
			&l.ItemCount, &l.DoneCount); err != nil {
			return nil, fmt.Errorf("scan shopping list: %w", err)
		}
		if catIDInt.Valid {
			l.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (db *DB) GetShoppingList(id int64) (*ShoppingList, error) {
	var l ShoppingList
	var catIDInt sql.NullInt64
	var catName, catColor, catIcon sql.NullString
	err := db.conn.QueryRow(`
		SELECT sl.id, sl.name, sl.category_id, sl.created_at, sl.updated_at,
			c.id, c.name, c.color, c.icon
		FROM shopping_lists sl
		LEFT JOIN categories c ON sl.category_id = c.id
		WHERE sl.id = ?`, id).
		Scan(&l.ID, &l.Name, &l.CategoryID, &l.CreatedAt, &l.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon)
	if err != nil {
		return nil, fmt.Errorf("get shopping list %d: %w", id, err)
	}
	if catIDInt.Valid {
		l.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
	}
	return &l, nil
}

func (db *DB) CreateShoppingList(l *ShoppingList) error {
	now := time.Now()
	res, err := db.conn.Exec("INSERT INTO shopping_lists (name, category_id, created_at, updated_at) VALUES (?, ?, ?, ?)",
		l.Name, l.CategoryID, now, now)
	if err != nil {
		return fmt.Errorf("insert shopping list: %w", err)
	}
	l.ID, _ = res.LastInsertId()
	l.CreatedAt = now
	l.UpdatedAt = now
	return nil
}

func (db *DB) UpdateShoppingList(l *ShoppingList) error {
	now := time.Now()
	_, err := db.conn.Exec("UPDATE shopping_lists SET name = ?, category_id = ?, updated_at = ? WHERE id = ?",
		l.Name, l.CategoryID, now, l.ID)
	if err != nil {
		return fmt.Errorf("update shopping list: %w", err)
	}
	l.UpdatedAt = now
	return nil
}

func (db *DB) DeleteShoppingList(id int64) error {
	_, err := db.conn.Exec("DELETE FROM shopping_lists WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete shopping list: %w", err)
	}
	return nil
}

func (db *DB) GetShoppingItems(listID int64, sort string) ([]ShoppingItem, error) {
	orderClause := "si.purchased ASC, si.sort_order ASC, si.created_at DESC"
	if sort == "name" {
		orderClause = "si.purchased ASC, LOWER(si.name) ASC"
	}
	rows, err := db.conn.Query(`
		SELECT si.id, si.list_id, si.name, si.quantity, si.unit, si.purchased,
			si.category_id, si.sort_order, si.created_at, si.updated_at,
			c.id, c.name, c.color, c.icon
		FROM shopping_items si
		LEFT JOIN categories c ON si.category_id = c.id
		WHERE si.list_id = ?
		ORDER BY `+orderClause, listID)
	if err != nil {
		return nil, fmt.Errorf("query shopping items: %w", err)
	}
	defer rows.Close()

	var items []ShoppingItem
	for rows.Next() {
		var item ShoppingItem
		var catIDInt sql.NullInt64
		var catName, catColor, catIcon sql.NullString
		if err := rows.Scan(&item.ID, &item.ListID, &item.Name, &item.Quantity, &item.Unit,
			&item.Purchased, &item.CategoryID, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon); err != nil {
			return nil, fmt.Errorf("scan shopping item: %w", err)
		}
		if catIDInt.Valid {
			item.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		tags, err := db.getShoppingItemTags(items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Tags = tags
	}

	return items, nil
}

func (db *DB) getShoppingItemTags(itemID int64) ([]Tag, error) {
	rows, err := db.conn.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN shopping_item_tags sit ON t.id = sit.tag_id
		WHERE sit.shopping_item_id = ?
		ORDER BY t.name`, itemID)
	if err != nil {
		return nil, fmt.Errorf("query shopping item tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (db *DB) GetShoppingItem(id int64) (*ShoppingItem, error) {
	var item ShoppingItem
	var catIDInt sql.NullInt64
	var catName, catColor, catIcon sql.NullString
	err := db.conn.QueryRow(`
		SELECT si.id, si.list_id, si.name, si.quantity, si.unit, si.purchased,
			si.category_id, si.sort_order, si.created_at, si.updated_at,
			c.id, c.name, c.color, c.icon
		FROM shopping_items si
		LEFT JOIN categories c ON si.category_id = c.id
		WHERE si.id = ?`, id).
		Scan(&item.ID, &item.ListID, &item.Name, &item.Quantity, &item.Unit,
			&item.Purchased, &item.CategoryID, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon)
	if err != nil {
		return nil, fmt.Errorf("get shopping item %d: %w", id, err)
	}
	if catIDInt.Valid {
		item.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
	}
	tags, err := db.getShoppingItemTags(item.ID)
	if err != nil {
		return nil, err
	}
	item.Tags = tags
	return &item, nil
}

func (db *DB) FindShoppingItemByName(listID int64, name string) (*ShoppingItem, error) {
	var item ShoppingItem
	var catIDInt sql.NullInt64
	var catName, catColor, catIcon sql.NullString
	err := db.conn.QueryRow(`
		SELECT si.id, si.list_id, si.name, si.quantity, si.unit, si.purchased,
			si.category_id, si.sort_order, si.created_at, si.updated_at,
			c.id, c.name, c.color, c.icon
		FROM shopping_items si
		LEFT JOIN categories c ON si.category_id = c.id
		WHERE si.list_id = ? AND LOWER(si.name) = LOWER(?)`, listID, name).
		Scan(&item.ID, &item.ListID, &item.Name, &item.Quantity, &item.Unit,
			&item.Purchased, &item.CategoryID, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon)
	if err != nil {
		return nil, err
	}
	if catIDInt.Valid {
		item.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
	}
	return &item, nil
}

func (db *DB) ReactivateShoppingItem(id int64, quantity int, unit string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE shopping_items SET purchased = 0, quantity = ?, unit = ?, updated_at = ? WHERE id = ?`,
		quantity, unit, now, id)
	if err != nil {
		return fmt.Errorf("reactivate shopping item: %w", err)
	}
	return nil
}

func (db *DB) CreateShoppingItem(item *ShoppingItem) error {
	now := time.Now()
	var maxOrder int
	db.conn.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM shopping_items WHERE list_id = ?", item.ListID).Scan(&maxOrder)

	res, err := db.conn.Exec(`
		INSERT INTO shopping_items (list_id, name, quantity, unit, purchased, category_id, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ListID, item.Name, item.Quantity, item.Unit, item.Purchased,
		item.CategoryID, maxOrder+1, now, now)
	if err != nil {
		return fmt.Errorf("insert shopping item: %w", err)
	}
	item.ID, _ = res.LastInsertId()
	item.SortOrder = maxOrder + 1
	item.CreatedAt = now
	item.UpdatedAt = now

	db.conn.Exec("UPDATE shopping_lists SET updated_at = ? WHERE id = ?", now, item.ListID)

	db.recordShoppingHistory(item.Name, item.Unit, item.Quantity, item.CategoryID)
	return nil
}

func (db *DB) UpdateShoppingItem(item *ShoppingItem) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE shopping_items SET name = ?, quantity = ?, unit = ?, purchased = ?,
			category_id = ?, sort_order = ?, updated_at = ?
		WHERE id = ?`,
		item.Name, item.Quantity, item.Unit, item.Purchased,
		item.CategoryID, item.SortOrder, now, item.ID)
	if err != nil {
		return fmt.Errorf("update shopping item: %w", err)
	}
	item.UpdatedAt = now

	db.conn.Exec("UPDATE shopping_lists SET updated_at = ? WHERE id = ?", now, item.ListID)
	return nil
}

func (db *DB) ToggleShoppingItem(id int64) error {
	_, err := db.conn.Exec(`
		UPDATE shopping_items SET purchased = NOT purchased, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("toggle shopping item: %w", err)
	}
	return nil
}

func (db *DB) DeleteShoppingItem(id int64) error {
	_, err := db.conn.Exec("DELETE FROM shopping_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete shopping item: %w", err)
	}
	return nil
}

func (db *DB) SetShoppingItemTags(itemID int64, tagIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM shopping_item_tags WHERE shopping_item_id = ?", itemID); err != nil {
		return fmt.Errorf("clear shopping item tags: %w", err)
	}

	if len(tagIDs) > 0 {
		placeholders := make([]string, len(tagIDs))
		args := make([]interface{}, 0, len(tagIDs)*2)
		for i, tid := range tagIDs {
			placeholders[i] = "(?, ?)"
			args = append(args, itemID, tid)
		}
		query := "INSERT INTO shopping_item_tags (shopping_item_id, tag_id) VALUES " + strings.Join(placeholders, ", ")
		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("insert shopping item tags: %w", err)
		}
	}

	return tx.Commit()
}

func (db *DB) recordShoppingHistory(name, unit string, quantity int, categoryID *int64) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	var existingID int64
	err := db.conn.QueryRow("SELECT id FROM shopping_history WHERE LOWER(name) = ?", normalized).Scan(&existingID)
	if err == nil {
		db.conn.Exec(`UPDATE shopping_history SET use_count = use_count + 1, last_used = CURRENT_TIMESTAMP,
			default_quantity = ?, unit = ?, category_id = ? WHERE id = ?`,
			quantity, unit, categoryID, existingID)
	} else {
		db.conn.Exec(`INSERT INTO shopping_history (name, unit, default_quantity, category_id, use_count, last_used)
			VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP)`,
			name, unit, quantity, categoryID)
	}
}

func (db *DB) SearchShoppingHistory(query string) ([]ShoppingHistory, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, unit, default_quantity, category_id, use_count, last_used
		FROM shopping_history
		WHERE LOWER(name) LIKE ?
		ORDER BY use_count DESC, last_used DESC
		LIMIT 20`, "%"+strings.ToLower(query)+"%")
	if err != nil {
		return nil, fmt.Errorf("search shopping history: %w", err)
	}
	defer rows.Close()

	var results []ShoppingHistory
	for rows.Next() {
		var h ShoppingHistory
		if err := rows.Scan(&h.ID, &h.Name, &h.Unit, &h.DefaultQuantity, &h.CategoryID, &h.UseCount, &h.LastUsed); err != nil {
			return nil, fmt.Errorf("scan shopping history: %w", err)
		}
		results = append(results, h)
	}
	return results, rows.Err()
}

func (db *DB) ReorderShoppingItems(listID int64, itemIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for i, id := range itemIDs {
		if _, err := tx.Exec("UPDATE shopping_items SET sort_order = ? WHERE id = ? AND list_id = ?", i, id, listID); err != nil {
			return fmt.Errorf("reorder shopping item: %w", err)
		}
	}

	return tx.Commit()
}
