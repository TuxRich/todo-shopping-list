# ToDo & Shopping List

A web-based todo and shopping list manager built in Go with SQLite. Designed to run as a Home Assistant Add-on with Ingress support.

## Features

**Todo Lists**
- Create multiple todo lists with categories
- Items with optional deadlines and date ranges (start/end)
- Tag items for easy filtering
- Drag-and-drop reordering
- Mark items complete/incomplete

**Shopping Lists**
- Create multiple shopping lists with categories
- Item quantity and unit tracking (e.g. "3 kg")
- Quick re-add from purchase history with autocomplete
- Mark items as purchased
- Item count shown on list overview

**General**
- Categories (name, color, icon) for organizing lists and items
- Tags (name, color) for labeling individual items
- Full REST API at `/api/v1/` for external integrations
- Home Assistant Add-on with Ingress support

## Running Locally

Requirements: Go 1.21+

```bash
go run .
```

The server starts on `http://localhost:8099`. The SQLite database is created at `./data/todo.db`.

Environment variables:
- `PORT` - Server port (default: `8099`)
- `DB_PATH` - SQLite database path (default: `./data/todo.db`)

## Home Assistant Add-on

### Installation

1. Copy this repository to your Home Assistant addons directory, or add it as a repository in the Add-on Store.
2. Install the "ToDo & Shopping List" add-on.
3. Start the add-on.
4. Access via the sidebar panel (Ingress).

### Manual Docker Build

```bash
docker build -t todo-app .
docker run -p 8099:8099 -v todo-data:/data todo-app
```

## REST API

All endpoints return JSON. Base path: `/api/v1/`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/todo-lists` | List all todo lists |
| POST | `/api/v1/todo-lists` | Create a todo list |
| GET | `/api/v1/todo-lists/:id` | Get a todo list |
| PUT | `/api/v1/todo-lists/:id` | Update a todo list |
| DELETE | `/api/v1/todo-lists/:id` | Delete a todo list |
| GET | `/api/v1/todo-lists/:id/items` | List items in a todo list |
| POST | `/api/v1/todo-lists/:id/items` | Add an item to a todo list |
| GET | `/api/v1/todo-items/:id` | Get a todo item |
| PUT | `/api/v1/todo-items/:id` | Update a todo item |
| DELETE | `/api/v1/todo-items/:id` | Delete a todo item |
| GET | `/api/v1/shopping-lists` | List all shopping lists |
| POST | `/api/v1/shopping-lists` | Create a shopping list |
| GET | `/api/v1/shopping-lists/:id` | Get a shopping list |
| PUT | `/api/v1/shopping-lists/:id` | Update a shopping list |
| DELETE | `/api/v1/shopping-lists/:id` | Delete a shopping list |
| GET | `/api/v1/shopping-lists/:id/items` | List items in a shopping list |
| POST | `/api/v1/shopping-lists/:id/items` | Add an item to a shopping list |
| GET | `/api/v1/shopping-items/:id` | Get a shopping item |
| PUT | `/api/v1/shopping-items/:id` | Update a shopping item |
| DELETE | `/api/v1/shopping-items/:id` | Delete a shopping item |
| GET | `/api/v1/shopping-history?q=...` | Search shopping history |
| GET | `/api/v1/categories` | List all categories |
| POST | `/api/v1/categories` | Create a category |
| PUT | `/api/v1/categories/:id` | Update a category |
| DELETE | `/api/v1/categories/:id` | Delete a category |
| GET | `/api/v1/tags` | List all tags |
| POST | `/api/v1/tags` | Create a tag |
| PUT | `/api/v1/tags/:id` | Update a tag |
| DELETE | `/api/v1/tags/:id` | Delete a tag |

## Tech Stack

- **Go** with chi router
- **SQLite** via modernc.org/sqlite (pure Go, no CGO)
- **HTMX** for interactive UI
- **Tailwind CSS** for styling
- **SortableJS** for drag-and-drop
