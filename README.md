# TuxRich Home Assistant Add-ons

## ToDo & Shopping List

A web-based todo and shopping list manager built in Go with SQLite, designed as a Home Assistant Add-on.

### Installation

1. In Home Assistant, go to **Settings > Add-ons > Add-on Store**.
2. Click the three-dot menu (top right) and select **Repositories**.
3. Add this repository URL: `https://github.com/TuxRich/todo-shopping-list`
4. Find "ToDo & Shopping List" in the store and click **Install**.
5. Start the add-on and open the **Web UI** from the sidebar.

### Features

**Todo Lists**
- Create multiple lists with categories
- Items with optional deadlines and date ranges
- Tag items for easy filtering
- Drag-and-drop reordering

**Shopping Lists**
- Item quantity and unit tracking
- Quick re-add from purchase history (autocomplete search)
- Duplicate detection: re-adding an existing item unchecks it instead
- Sort by name or manual order
- Item count shown on list overview

**General**
- Categories (name, color, icon) and tags (name, color)
- Full REST API at `/api/v1/` for external integrations
- Home Assistant Ingress support (embedded in sidebar)

### Running Standalone (without Home Assistant)

```bash
cd todo-shopping-list
go run .
```

The server starts on `http://localhost:8099`. The SQLite database is created at `./data/todo.db`.

Environment variables:
- `PORT` - Server port (default: `8099`)
- `DB_PATH` - SQLite database path (default: `./data/todo.db`)

### REST API

All endpoints return JSON. Base path: `/api/v1/`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/POST | `/api/v1/todo-lists` | List / create todo lists |
| GET/PUT/DELETE | `/api/v1/todo-lists/:id` | Get / update / delete a todo list |
| GET/POST | `/api/v1/todo-lists/:id/items` | List / add todo items |
| GET/PUT/DELETE | `/api/v1/todo-items/:id` | Get / update / delete a todo item |
| GET/POST | `/api/v1/shopping-lists` | List / create shopping lists |
| GET/PUT/DELETE | `/api/v1/shopping-lists/:id` | Get / update / delete a shopping list |
| GET/POST | `/api/v1/shopping-lists/:id/items` | List / add shopping items |
| GET/PUT/DELETE | `/api/v1/shopping-items/:id` | Get / update / delete a shopping item |
| GET | `/api/v1/shopping-history?q=...` | Search shopping history |
| GET/POST | `/api/v1/categories` | List / create categories |
| GET/PUT/DELETE | `/api/v1/categories/:id` | Get / update / delete a category |
| GET/POST | `/api/v1/tags` | List / create tags |
| GET/PUT/DELETE | `/api/v1/tags/:id` | Get / update / delete a tag |
