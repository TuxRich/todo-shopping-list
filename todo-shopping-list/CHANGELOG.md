## 1.2.1

- Fixed dragging on the board selecting text instead of moving the card: after an update, a browser or a caching proxy in front of Home Assistant could keep using the previous version's scripts. Script and style links now change with every release, so the current version is always loaded
- Board cards can no longer be text-selected, and on touch screens a short press-and-hold picks a card up while a quick swipe still scrolls the page

## 1.2.0

- Tasks can now be marked To do, In progress, or Done; the list view shows an "In progress" badge and has a start/pause button on each task
- Added a board view for todo lists: To do, In progress, and Done columns side by side, with drag and drop between them and move buttons on each card
- Existing tasks are upgraded automatically on first start: completed tasks become Done and the rest To do
- REST API: todo items have a new `status` field (`todo`, `in_progress`, `done`); the `completed` flag is still accepted and kept in sync
- Fixed drag-and-drop reordering not working after a page refresh or when opening a list directly, for both todo lists and shopping lists in manual sort

## 1.1.2

- Fixed saving an item together with its tags only half-succeeding when one part failed; the item and its tags are now saved together or not at all, both when adding and when editing
- Ticking or unticking an item now counts as activity, moving its list to the top of the recent lists
- Fixed re-adding an existing shopping item through the REST API ignoring errors and returning an empty response
- Fixed a database error during duplicate detection causing a second copy of a shopping item to be created
- Invalid tag values in a form submission are now ignored instead of failing the whole save

## 1.1.1

- Fixed a cross-site scripting hole: a category or tag name containing HTML ran as code when its edit dialog was opened
- Fixed REST API `PUT` requests wiping every field left out of the request body (quantity, unit, deadlines, colors, sort order); updates now change only the fields you send
- REST API `PUT` now accepts `"category_id": null` to clear a category, and leaves it unchanged when the field is omitted
- Fixed re-adding an existing shopping item clearing its unit; re-adding now also updates the list timestamp and the item's history count
- Fixed a server error when setting tags on an item that no longer exists (now returns "not found")
- Blank or whitespace-only list, item, category, and tag names are now rejected instead of being saved
- Failed database writes now report an error instead of silently appearing to succeed

## 1.1.0

- Shopping lists now sort by name by default instead of manual order
- Added a Category sort option for shopping lists (groups items by category)

## 1.0.3

- Exposed the REST API on host port 8099 for external integrations (e.g. Home Assistant automations)

## 1.0.2

- Catppuccin Mocha dark theme to match Home Assistant styling
- Fixed all navigation links when running inside Home Assistant Ingress
- Fixed HTTP redirects to respect Ingress base path
- Made all URLs (templates, HTMX, fetch, static assets) relative for proper Ingress support

## 1.0.1

- Fixed Docker build: updated Go builder image to 1.25-alpine to match go.mod
- Fixed s6-overlay startup error by adding `init: false` to config
- Added default value for BUILD_FROM arg in Dockerfile

## 1.0.0

- Initial release
- Todo lists with deadlines, date ranges, and tags
- Shopping lists with quantity, history search, and duplicate prevention
- Categories and tags management
- Sort shopping lists by name
- REST API for external integrations
- Home Assistant Ingress support
