# Groupie Tracker Geolocalization

## Overview
Groupie Tracker is a web application that consumes a RESTful API to display information about various musical artists and bands. It provides a user-friendly interface to view artist details, their concert locations, and dates. 

This specific project version focuses on the **Geolocalization**, **Search-Bar** and **Filters** features. It allows users to narrow down the list of artists based on various criteria, and provides a rich, interactive map to visualize the chronological path of an artist's tour across the globe.

## Features
- **Artist Display:** View a grid of artists with their images, names, and creation dates.
- **Geolocalization & Mapping:** 
  - Interactive Leaflet map displaying concert locations.
  - Client-side geocoding utilizing the OpenStreetMap Nominatim API.
  - Progressive chronological animation showing the tour path drawing live.
  - Intelligent location correction map to correctly resolve ambiguous geographic names.
- **Filtering System:**
  - Creation Date (Min/Max Year)
  - First Album Year (Min/Max Year)
  - Number of Members (Checkboxes)
  - Concert Locations (Text match)
- **Live Search Functionality:** Real-time, asynchronous search bar. Includes autocomplete suggestions triggered at 3+ characters that display the match type (artist/band, member, location, first album date, or creation date) and redirect directly to the artist details page.

## Project Structure
- `cmd/`: Contains the application entry point (`main.go`).
- `internal/`: Contains the core backend logic (`server.go`, `handlers.go`, `models.go`, `apifetch.go`, `filters.go`).
- `templates/`: Contains HTML files for the frontend views (`index.html`, `artist.html`, `error.html`).
- `static/`: Contains static assets like CSS (`style.css`) and JavaScript files (`js.js`, `map.js`).
- `docs/`: Project documentation and guidelines.

## Prerequisites
- Go (version 1.18 or higher recommended)

## How to Run
1. Clone this repository.
2. Navigate to the project root directory:
   ```bash
   cd groupie-tracker-geolocalization
   ```
3. Run the application:
   ```bash
   go run cmd/main.go
   ```
4. Open your web browser and navigate to `http://localhost:8080`.
