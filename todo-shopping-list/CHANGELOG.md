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
