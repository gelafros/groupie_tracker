# Groupie Tracker Visualizations

## Overview
Groupie Tracker is a web application that consumes a RESTful API to display information about various musical artists and bands. It provides a premium, visually polished interface designed around **Schneiderman's 8 Golden Rules of Interface Design**.

This project version focuses on **Visualizations** — presenting the API data in the most attractive and usable way possible through modern CSS, responsive layouts, smooth animations, and interactive elements.

## Schneiderman's 8 Golden Rules

| Rule | Implementation |
|------|---------------|
| **Consistency** | Unified design tokens (colors, typography, spacing) across all pages via CSS custom properties. |
| **Shortcuts** | Press `/` to focus the search bar instantly from anywhere on the page. |
| **Informative Feedback** | Live result counter, loading states, and descriptive error messages. |
| **Dialogue Closure** | Clear "Apply Filters" / "Reset" actions. Suggestions navigate directly to artist pages. |
| **Error Handling** | Styled error pages for 404, 400, and 500 with clear messages and a "Go Home" button. |
| **Easy Reversal** | "Reset" clears all filters. "← Back to Artists" on detail pages. Clickable logo returns home. |
| **Internal Locus of Control** | Users control search scope, filter ranges, member counts, and location filters freely. |
| **Reduce Memory Load** | All info visible on cards. Artist details, members, map, and timeline on one page. |

## Features
- **Premium Dark Theme:** Animated gradient background, glassmorphism effects, Inter font, and accent glow animations.
- **Artist Display:** Responsive grid of artists with hover glow effects and staggered fade-in animations.
- **Geolocalization & Mapping:**
  - Interactive Leaflet map displaying concert locations.
  - Client-side geocoding via the OpenStreetMap Nominatim API.
  - Progressive chronological animation showing the tour path drawing live.
  - Chronological timeline sidebar alongside the map.
- **Filtering System:**
  - Creation Date (Min/Max Year)
  - First Album Year (Min/Max Year)
  - Number of Members (Checkboxes)
  - Concert Locations (Text match)
- **Live Search:** Real-time async search with autocomplete suggestions showing match type and band context.
- **Keyboard Shortcuts:** Press `/` to focus search.
- **Error Pages:** Custom-styled 404/400/500 pages with glowing animated error codes.
- **Responsive Design:** 3 breakpoints (420px, 768px, 1024px) ensuring consistency across all screen sizes.
- **SEO:** Meta descriptions, semantic HTML, and proper heading hierarchy on every page.

## Project Structure
- `cmd/` — Application entry point (`main.go`).
- `internal/` — Core backend logic (`server.go`, `handlers.go`, `models.go`, `apifetch.go`, `filters.go`, `handlers_test.go`).
- `templates/` — HTML templates (`index.html`, `artist.html`, `error.html`).
- `static/css/` — Stylesheet (`style.css`).
- `static/javascript/` — Client-side logic (`js.js`, `map.js`).

## Prerequisites
- Go (version 1.18 or higher)

## How to Run
1. Clone this repository.
2. Navigate to the project root:
   ```bash
   cd groupie-tracker-visualizations
   ```
3. Run the application:
   ```bash
   go run cmd/main.go
   ```
4. Open `http://localhost:8080` in your browser.

## How to Test
```bash
go test ./internal/ -v
```

Runs 20 tests covering filters, handlers, search API, suggestions API, and error rendering.
