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
