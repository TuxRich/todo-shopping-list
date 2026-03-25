package database

import "fmt"

func (db *DB) GetCategories() ([]Category, error) {
	rows, err := db.conn.Query("SELECT id, name, color, icon FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.Icon); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (db *DB) GetCategory(id int64) (*Category, error) {
	var c Category
	err := db.conn.QueryRow("SELECT id, name, color, icon FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Color, &c.Icon)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", id, err)
	}
	return &c, nil
}

func (db *DB) CreateCategory(c *Category) error {
	res, err := db.conn.Exec("INSERT INTO categories (name, color, icon) VALUES (?, ?, ?)",
		c.Name, c.Color, c.Icon)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) UpdateCategory(c *Category) error {
	_, err := db.conn.Exec("UPDATE categories SET name = ?, color = ?, icon = ? WHERE id = ?",
		c.Name, c.Color, c.Icon, c.ID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (db *DB) DeleteCategory(id int64) error {
	_, err := db.conn.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}
