# Groupie Tracker Filters

## Overview
Groupie Tracker is a web application that consumes a RESTful API to display information about various musical artists and bands. It provides a user-friendly interface to view artist details, their concert locations, and dates. 

This specific project branch focuses on the **Filters** feature, allowing users to narrow down the list of artists based on various criteria.

## Features
- **Artist Display:** View a grid of artists with their images, names, and creation dates.
- **Artist Details:** Detailed view for each artist showing members, first album date, and a relation of concert locations to dates.
- **Filtering System:**
  - Creation Date (Min/Max Year)
  - First Album Year (Min/Max Year)
  - Number of Members (Checkboxes)
  - Concert Locations (Text match)
- **Search Functionality:** Client-side search bar to quickly find artists by name (Note: Backend endpoint currently pending implementation).

## Project Structure
- `cmd/`: Contains the application entry point (`main.go`).
- `internal/`: Contains the core backend logic (`server.go`, `handlers.go`, `models.go`, `apifetch.go`, `filters.go`).
- `templates/`: Contains HTML files for the frontend views.
- `static/`: Contains static assets like CSS and JavaScript files.
- `docs/`: Project documentation and guidelines.

## Prerequisites
- Go (version 1.18 or higher recommended)

## How to Run
1. Clone this repository.
2. Navigate to the project root directory:
   ```bash
   cd groupie-tracker-filters
   ```
3. Run the application:
   ```bash
   go run cmd/main.go
   ```
4. Open your web browser and navigate to `http://localhost:8080`.
