package internal

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// renderError is a helper function to render a user-friendly error page.
func renderError(w http.ResponseWriter, status int, title, message string) {
	w.WriteHeader(status)
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, message, status)
		return
	}
	data := struct {
		Status  int
		Title   string
		Message string
	}{status, title, message}
	tmpl.Execute(w, data)
}

// homeHandler parses and serves the index.html template for the root route.
// It handles template parsing and execution errors by responding with a 500 status.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that only the exact root path "/" is handled here to prevent it from acting as a catch-all
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "Not Found", "The page you are looking for does not exist.")
		return
	}

	// Parse filter queries from the request URL
	filters := ParseFilters(r)

	// bands slice will hold the final filtered artists
	var bands []BandInfo

	for _, artist := range cachedArtists {
		// 1. Filter by Creation Date
		if artist.CreationDate < filters.CreationMin || artist.CreationDate > filters.CreationMax {
			continue
		}
		
		// 2. Filter by First Album Year (Format: DD-MM-YYYY)
		albumParts := strings.Split(artist.FirstAlbum, "-")
		if len(albumParts) == 3 {
			albumYear, err := strconv.Atoi(albumParts[2])
			if err == nil {
				if albumYear < filters.FirstAlbumMin || albumYear > filters.FirstAlbumMax {
					continue
				}
			}
		}
		
		// 3. Filter by Number of Members
		if len(filters.Members) > 0 {
			matchedMembers := false
			memberCountStr := strconv.Itoa(len(artist.Members))
			for _, m := range filters.Members {
				if m == memberCountStr {
					matchedMembers = true
					break
				}
			}
			if !matchedMembers {
				continue
			}
		}

		// 4. Filter by Concert Location
		artistLocs, locExists := cachedLocationMap[artist.ID]
		if filters.Location != "" {
			matchedLocation := false
			if locExists {
				for _, loc := range artistLocs.Locations {
					locClean := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(loc, "_", " "), "-", " "))
					if strings.Contains(locClean, filters.Location) {
						matchedLocation = true
						break
					}
				}
			}
			if !matchedLocation {
				continue
			}
		}

		var formattedLocations []string
		if locExists {
			for _, location := range artistLocs.Locations {
				cleanLoc := strings.ReplaceAll(location, "_", " ")
				cleanLoc = strings.ReplaceAll(cleanLoc, "-", " ")
				formattedLocations = append(formattedLocations, cleanLoc)
			}
		}

		formattedRelations := make(map[string][]string)
		artistRels, relExists := cachedRelationsMap[artist.ID]
		if relExists {
			for loc, datesList := range artistRels.DatesLocations {
				cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
				cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
				formattedRelations[cleanRelLoc] = datesList
			}
		}

		var formattedDates []string
		artistDates, dateExists := cachedDatesMap[artist.ID]
		if dateExists {
			for _, date := range artistDates.Dates {
				cleanDate := strings.ReplaceAll(date, "*", "")
				formattedDates = append(formattedDates, cleanDate)
			}
		}

		band := BandInfo{
			ID:           artist.ID,
			Name:         artist.Name,
			CreationDate: artist.CreationDate,
			Image:        artist.Image,
			Locations:    formattedLocations,
			Dates:        formattedDates,
			Relations:    formattedRelations,
			Members:      artist.Members,
			FirstAlbum:   artist.FirstAlbum,
		}

		bands = append(bands, band)
	}

	// Execute the parsed template, passing the fetched artist data to it
	data := struct {
		Title   string
		Artists []BandInfo
	}{
		Title:   "Groupie Tracker",
		Artists: bands,
	}

	output, err := template.ParseFiles("templates/index.html")
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not parse template.")
		return
	}

	err = output.Execute(w, data)
	// Handle any errors that occur during template execution
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution failed.")
		return
	}
}

// artistHandler handles the individual artist details page.
func artistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		renderError(w, http.StatusBadRequest, "Bad Request", "Invalid artist ID provided.")
		return
	}

	var foundBand *BandInfo
	for _, artist := range cachedArtists {
		if artist.ID == id {
			formattedRelations := make(map[string][]string)

			// Match relation by ID safely
			if rel, exists := cachedRelationsMap[artist.ID]; exists {
				for loc, datesList := range rel.DatesLocations {
					cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
					cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
					formattedRelations[cleanRelLoc] = datesList
				}
			}

			foundBand = &BandInfo{
				ID:           artist.ID,
				Name:         artist.Name,
				CreationDate: artist.CreationDate,
				Image:        artist.Image,
				Relations:    formattedRelations,
				Members:      artist.Members,
				FirstAlbum:   artist.FirstAlbum,
			}
			break
		}
	}

	if foundBand == nil {
		renderError(w, http.StatusNotFound, "Not Found", "Artist not found.")
		return
	}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not load artist template.")
		return
	}

	data := struct {
		Title  string
		Artist *BandInfo
	}{
		Title:  foundBand.Name + " - Groupie Tracker",
		Artist: foundBand,
	}

	if err := tmpl.Execute(w, data); err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution failed.")
	}
}

// searchHandler handles asynchronous requests for filtering and searching.
// It returns a JSON array of artists matching all selected criteria.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse standard filters (year ranges, members, locations)
	filters := ParseFilters(r)
	
	// Grab the search query for the artist name
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	var filteredArtists []Artist

	for _, artist := range cachedArtists {
		// Apply name search
		if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) {
			continue
		}

		// 1. Filter by Creation Date
		if artist.CreationDate < filters.CreationMin || artist.CreationDate > filters.CreationMax {
			continue
		}
		
		// 2. Filter by First Album Year (Format: DD-MM-YYYY)
		albumParts := strings.Split(artist.FirstAlbum, "-")
		if len(albumParts) == 3 {
			albumYear, err := strconv.Atoi(albumParts[2])
			if err == nil {
				if albumYear < filters.FirstAlbumMin || albumYear > filters.FirstAlbumMax {
					continue
				}
			}
		}
		
		// 3. Filter by Number of Members
		if len(filters.Members) > 0 {
			matchedMembers := false
			memberCountStr := strconv.Itoa(len(artist.Members))
			for _, m := range filters.Members {
				if m == memberCountStr {
					matchedMembers = true
					break
				}
			}
			if !matchedMembers {
				continue
			}
		}

		// 4. Filter by Concert Location
		artistLocs, locExists := cachedLocationMap[artist.ID]
		if filters.Location != "" {
			matchedLocation := false
			if locExists {
				for _, loc := range artistLocs.Locations {
					locClean := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(loc, "_", " "), "-", " "))
					if strings.Contains(locClean, filters.Location) {
						matchedLocation = true
						break
					}
				}
			}
			if !matchedLocation {
				continue
			}
		}

		filteredArtists = append(filteredArtists, artist)
	}

	// Ensure an empty slice is returned as [] instead of null in JSON
	if filteredArtists == nil {
		filteredArtists = []Artist{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredArtists); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
