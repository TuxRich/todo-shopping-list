package database

import "fmt"

func (db *DB) GetTags() ([]Tag, error) {
	rows, err := db.conn.Query("SELECT id, name, color FROM tags ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
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

func (db *DB) GetTag(id int64) (*Tag, error) {
	var t Tag
	err := db.conn.QueryRow("SELECT id, name, color FROM tags WHERE id = ?", id).
		Scan(&t.ID, &t.Name, &t.Color)
	if err != nil {
		return nil, fmt.Errorf("get tag %d: %w", id, err)
	}
	return &t, nil
}

func (db *DB) CreateTag(t *Tag) error {
	res, err := db.conn.Exec("INSERT INTO tags (name, color) VALUES (?, ?)", t.Name, t.Color)
	if err != nil {
		return fmt.Errorf("insert tag: %w", err)
	}
	t.ID, _ = res.LastInsertId()
	return nil
}

func (db *DB) UpdateTag(t *Tag) error {
	_, err := db.conn.Exec("UPDATE tags SET name = ?, color = ? WHERE id = ?", t.Name, t.Color, t.ID)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}
	return nil
}

func (db *DB) DeleteTag(id int64) error {
	_, err := db.conn.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}
