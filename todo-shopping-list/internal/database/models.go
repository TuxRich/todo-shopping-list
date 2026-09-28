package database

import "time"

type Category struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type TodoList struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	CategoryID *int64    `json:"category_id,omitempty"`
	Category   *Category `json:"category,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ItemCount  int       `json:"item_count"`
	DoneCount  int       `json:"done_count"`
}

// Todo item statuses. Status is the source of truth; Completed mirrors
// Status == TodoDone for the list counts and for API clients that predate it.
const (
	TodoTodo       = "todo"
	TodoInProgress = "in_progress"
	TodoDone       = "done"
)

// ValidTodoStatus reports whether s is one of the known todo item statuses.
func ValidTodoStatus(s string) bool {
	return s == TodoTodo || s == TodoInProgress || s == TodoDone
}

type TodoItem struct {
	ID          int64     `json:"id"`
	ListID      int64     `json:"list_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Completed   bool      `json:"completed"`
	Deadline    *string   `json:"deadline,omitempty"`
	DateStart   *string   `json:"date_start,omitempty"`
	DateEnd     *string   `json:"date_end,omitempty"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []Tag     `json:"tags,omitempty"`
}

// SetCompleted applies a legacy completed flag. True moves the item to done;
// false reopens a done item as todo but leaves an in-progress item where it is.
func (item *TodoItem) SetCompleted(completed bool) {
	switch {
	case completed:
		item.Status = TodoDone
	case item.Status == TodoDone || item.Status == "":
		item.Status = TodoTodo
	}
	item.Completed = item.Status == TodoDone
}

type ShoppingList struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	CategoryID *int64    `json:"category_id,omitempty"`
	Category   *Category `json:"category,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ItemCount  int       `json:"item_count"`
	DoneCount  int       `json:"done_count"`
}

type ShoppingItem struct {
	ID         int64     `json:"id"`
	ListID     int64     `json:"list_id"`
	Name       string    `json:"name"`
	Quantity   int       `json:"quantity"`
	Unit       string    `json:"unit"`
	Purchased  bool      `json:"purchased"`
	CategoryID *int64    `json:"category_id,omitempty"`
	Category   *Category `json:"category,omitempty"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Tags       []Tag     `json:"tags,omitempty"`
}

type ShoppingHistory struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Unit            string    `json:"unit"`
	DefaultQuantity int       `json:"default_quantity"`
	CategoryID      *int64    `json:"category_id,omitempty"`
	UseCount        int       `json:"use_count"`
	LastUsed        time.Time `json:"last_used"`
}
