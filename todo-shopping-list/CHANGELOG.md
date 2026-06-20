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
