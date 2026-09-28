package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	conn.SetMaxOpenConns(1)

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// querier is satisfied by both *sql.DB and *sql.Tx, so the same statement
// helpers can run standalone or inside a transaction.
type querier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// inTx runs fn in a transaction, committing only if it returns nil. The pool
// holds a single connection, so fn must use tx and never db.conn, or it will
// block forever waiting for the connection the transaction already holds.
func (db *DB) inTx(fn func(tx *sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT NOT NULL DEFAULT '#6366f1',
			icon TEXT NOT NULL DEFAULT '📁'
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			color TEXT NOT NULL DEFAULT '#8b5cf6'
		)`,
		`CREATE TABLE IF NOT EXISTS todo_lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS todo_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id INTEGER NOT NULL REFERENCES todo_lists(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed BOOLEAN NOT NULL DEFAULT 0,
			deadline TEXT,
			date_start TEXT,
			date_end TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS todo_item_tags (
			todo_item_id INTEGER NOT NULL REFERENCES todo_items(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (todo_item_id, tag_id)
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 1,
			unit TEXT NOT NULL DEFAULT '',
			purchased BOOLEAN NOT NULL DEFAULT 0,
			category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_item_tags (
			shopping_item_id INTEGER NOT NULL REFERENCES shopping_items(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (shopping_item_id, tag_id)
		)`,
		`CREATE TABLE IF NOT EXISTS shopping_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			unit TEXT NOT NULL DEFAULT '',
			default_quantity INTEGER NOT NULL DEFAULT 1,
			category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
			use_count INTEGER NOT NULL DEFAULT 1,
			last_used DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w\nSQL: %s", err, m)
		}
	}

	return db.addTodoStatus()
}

// addTodoStatus adds the todo_items.status column to databases created before
// statuses existed, backfilling it from completed. It is a no-op once the
// column is present, so it is safe to run on every start.
func (db *DB) addTodoStatus() error {
	var exists int
	if err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('todo_items') WHERE name = 'status'").Scan(&exists); err != nil {
		return fmt.Errorf("inspect todo_items: %w", err)
	}
	if exists > 0 {
		return nil
	}

	return db.inTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`ALTER TABLE todo_items ADD COLUMN status TEXT NOT NULL DEFAULT 'todo'
			CHECK (status IN ('todo', 'in_progress', 'done'))`); err != nil {
			return fmt.Errorf("add todo status column: %w", err)
		}
		if _, err := tx.Exec("UPDATE todo_items SET status = 'done' WHERE completed = 1"); err != nil {
			return fmt.Errorf("backfill todo status: %w", err)
		}
		return nil
	})
}
