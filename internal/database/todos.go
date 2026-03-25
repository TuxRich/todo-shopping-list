package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (db *DB) GetTodoLists() ([]TodoList, error) {
	rows, err := db.conn.Query(`
		SELECT tl.id, tl.name, tl.category_id, tl.created_at, tl.updated_at,
			c.id, c.name, c.color, c.icon,
			COUNT(ti.id) AS item_count,
			COUNT(CASE WHEN ti.completed = 1 THEN 1 END) AS done_count
		FROM todo_lists tl
		LEFT JOIN categories c ON tl.category_id = c.id
		LEFT JOIN todo_items ti ON ti.list_id = tl.id
		GROUP BY tl.id
		ORDER BY tl.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query todo lists: %w", err)
	}
	defer rows.Close()

	var lists []TodoList
	for rows.Next() {
		var l TodoList
		var catID, catName, catColor, catIcon sql.NullString
		var catIDInt sql.NullInt64
		if err := rows.Scan(&l.ID, &l.Name, &l.CategoryID, &l.CreatedAt, &l.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon,
			&l.ItemCount, &l.DoneCount); err != nil {
			return nil, fmt.Errorf("scan todo list: %w", err)
		}
		_ = catID
		if catIDInt.Valid {
			l.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (db *DB) GetTodoList(id int64) (*TodoList, error) {
	var l TodoList
	var catIDInt sql.NullInt64
	var catName, catColor, catIcon sql.NullString
	err := db.conn.QueryRow(`
		SELECT tl.id, tl.name, tl.category_id, tl.created_at, tl.updated_at,
			c.id, c.name, c.color, c.icon
		FROM todo_lists tl
		LEFT JOIN categories c ON tl.category_id = c.id
		WHERE tl.id = ?`, id).
		Scan(&l.ID, &l.Name, &l.CategoryID, &l.CreatedAt, &l.UpdatedAt,
			&catIDInt, &catName, &catColor, &catIcon)
	if err != nil {
		return nil, fmt.Errorf("get todo list %d: %w", id, err)
	}
	if catIDInt.Valid {
		l.Category = &Category{ID: catIDInt.Int64, Name: catName.String, Color: catColor.String, Icon: catIcon.String}
	}
	return &l, nil
}

func (db *DB) CreateTodoList(l *TodoList) error {
	now := time.Now()
	res, err := db.conn.Exec("INSERT INTO todo_lists (name, category_id, created_at, updated_at) VALUES (?, ?, ?, ?)",
		l.Name, l.CategoryID, now, now)
	if err != nil {
		return fmt.Errorf("insert todo list: %w", err)
	}
	l.ID, _ = res.LastInsertId()
	l.CreatedAt = now
	l.UpdatedAt = now
	return nil
}

func (db *DB) UpdateTodoList(l *TodoList) error {
	now := time.Now()
	_, err := db.conn.Exec("UPDATE todo_lists SET name = ?, category_id = ?, updated_at = ? WHERE id = ?",
		l.Name, l.CategoryID, now, l.ID)
	if err != nil {
		return fmt.Errorf("update todo list: %w", err)
	}
	l.UpdatedAt = now
	return nil
}

func (db *DB) DeleteTodoList(id int64) error {
	_, err := db.conn.Exec("DELETE FROM todo_lists WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete todo list: %w", err)
	}
	return nil
}

func (db *DB) GetTodoItems(listID int64) ([]TodoItem, error) {
	rows, err := db.conn.Query(`
		SELECT id, list_id, title, description, completed, deadline, date_start, date_end,
			sort_order, created_at, updated_at
		FROM todo_items WHERE list_id = ?
		ORDER BY completed ASC, sort_order ASC, created_at DESC`, listID)
	if err != nil {
		return nil, fmt.Errorf("query todo items: %w", err)
	}
	defer rows.Close()

	var items []TodoItem
	for rows.Next() {
		var item TodoItem
		if err := rows.Scan(&item.ID, &item.ListID, &item.Title, &item.Description,
			&item.Completed, &item.Deadline, &item.DateStart, &item.DateEnd,
			&item.SortOrder, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan todo item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		tags, err := db.getTodoItemTags(items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Tags = tags
	}

	return items, nil
}

func (db *DB) getTodoItemTags(itemID int64) ([]Tag, error) {
	rows, err := db.conn.Query(`
		SELECT t.id, t.name, t.color
		FROM tags t
		JOIN todo_item_tags tit ON t.id = tit.tag_id
		WHERE tit.todo_item_id = ?
		ORDER BY t.name`, itemID)
	if err != nil {
		return nil, fmt.Errorf("query todo item tags: %w", err)
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

func (db *DB) GetTodoItem(id int64) (*TodoItem, error) {
	var item TodoItem
	err := db.conn.QueryRow(`
		SELECT id, list_id, title, description, completed, deadline, date_start, date_end,
			sort_order, created_at, updated_at
		FROM todo_items WHERE id = ?`, id).
		Scan(&item.ID, &item.ListID, &item.Title, &item.Description,
			&item.Completed, &item.Deadline, &item.DateStart, &item.DateEnd,
			&item.SortOrder, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get todo item %d: %w", id, err)
	}
	tags, err := db.getTodoItemTags(item.ID)
	if err != nil {
		return nil, err
	}
	item.Tags = tags
	return &item, nil
}

func (db *DB) CreateTodoItem(item *TodoItem) error {
	now := time.Now()
	var maxOrder int
	db.conn.QueryRow("SELECT COALESCE(MAX(sort_order), 0) FROM todo_items WHERE list_id = ?", item.ListID).Scan(&maxOrder)

	res, err := db.conn.Exec(`
		INSERT INTO todo_items (list_id, title, description, completed, deadline, date_start, date_end, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ListID, item.Title, item.Description, item.Completed,
		item.Deadline, item.DateStart, item.DateEnd, maxOrder+1, now, now)
	if err != nil {
		return fmt.Errorf("insert todo item: %w", err)
	}
	item.ID, _ = res.LastInsertId()
	item.SortOrder = maxOrder + 1
	item.CreatedAt = now
	item.UpdatedAt = now

	db.conn.Exec("UPDATE todo_lists SET updated_at = ? WHERE id = ?", now, item.ListID)
	return nil
}

func (db *DB) UpdateTodoItem(item *TodoItem) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE todo_items SET title = ?, description = ?, completed = ?, deadline = ?,
			date_start = ?, date_end = ?, sort_order = ?, updated_at = ?
		WHERE id = ?`,
		item.Title, item.Description, item.Completed,
		item.Deadline, item.DateStart, item.DateEnd,
		item.SortOrder, now, item.ID)
	if err != nil {
		return fmt.Errorf("update todo item: %w", err)
	}
	item.UpdatedAt = now

	db.conn.Exec("UPDATE todo_lists SET updated_at = ? WHERE id = ?", now, item.ListID)
	return nil
}

func (db *DB) ToggleTodoItem(id int64) error {
	_, err := db.conn.Exec(`
		UPDATE todo_items SET completed = NOT completed, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("toggle todo item: %w", err)
	}
	return nil
}

func (db *DB) DeleteTodoItem(id int64) error {
	_, err := db.conn.Exec("DELETE FROM todo_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete todo item: %w", err)
	}
	return nil
}

func (db *DB) SetTodoItemTags(itemID int64, tagIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM todo_item_tags WHERE todo_item_id = ?", itemID); err != nil {
		return fmt.Errorf("clear todo item tags: %w", err)
	}

	if len(tagIDs) > 0 {
		placeholders := make([]string, len(tagIDs))
		args := make([]interface{}, 0, len(tagIDs)*2)
		for i, tid := range tagIDs {
			placeholders[i] = "(?, ?)"
			args = append(args, itemID, tid)
		}
		query := "INSERT INTO todo_item_tags (todo_item_id, tag_id) VALUES " + strings.Join(placeholders, ", ")
		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("insert todo item tags: %w", err)
		}
	}

	return tx.Commit()
}

func (db *DB) ReorderTodoItems(listID int64, itemIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for i, id := range itemIDs {
		if _, err := tx.Exec("UPDATE todo_items SET sort_order = ? WHERE id = ? AND list_id = ?", i, id, listID); err != nil {
			return fmt.Errorf("reorder todo item: %w", err)
		}
	}

	return tx.Commit()
}
