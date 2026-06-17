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
// It applies user-selected filters on the pre-processed cache to return matching bands.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that only the exact root path "/" is handled here to prevent it from acting as a catch-all
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "Not Found", "The page you are looking for does not exist.")
		return
	}

	// Parse filter queries from the request URL (creationDateMin/Max, firstAlbumMin/Max, members, location)
	filters := ParseFilters(r)

	// bands slice will hold the final filtered list of BandInfo to render in the HTML template
	var bands []BandInfo

	// Loop over our pre-processed artists to filter them efficiently without string splits or allocations
	for _, pa := range processedArtists {
		artist := pa.Artist

		// 1. Filter by Creation Date (numeric comparison)
		if artist.CreationDate < filters.CreationMin || artist.CreationDate > filters.CreationMax {
			continue
		}

		// 2. Filter by First Album Year (uses pre-parsed FirstAlbumYear integer)
		if pa.FirstAlbumYear < filters.FirstAlbumMin || pa.FirstAlbumYear > filters.FirstAlbumMax {
			continue
		}

		// 3. Filter by Number of Members
		// If member filter options are selected (e.g. 1, 2, 4), verify if the artist's member count is matched.
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
		// Performs a fast check against our pre-cleaned locations slice
		if filters.Location != "" {
			matchedLocation := false
			for _, locClean := range pa.CleanLocations {
				if strings.Contains(locClean, filters.Location) {
					matchedLocation = true
					break
				}
			}
			if !matchedLocation {
				continue
			}
		}

		// Since this artist passed all filters, append the pre-formatted BandInfo representation
		bands = append(bands, pa.BandInfo)
	}

	// Prepare data structure for the template execution
	data := struct {
		Title   string
		Artists []BandInfo
	}{
		Title:   "Groupie Tracker",
		Artists: bands,
	}

	// Load and parse the index dashboard template
	output, err := template.ParseFiles("templates/index.html")
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not parse template.")
		return
	}

	// Render the template with the filtered bands
	err = output.Execute(w, data)
	// Handle any errors that occur during template execution
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution failed.")
		return
	}
}

// artistHandler handles the individual artist details page.
// It retrieves the artist ID from query parameters and finds the corresponding pre-formatted BandInfo.
func artistHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameter "id"
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		renderError(w, http.StatusBadRequest, "Bad Request", "Invalid artist ID provided.")
		return
	}

	// Find the matching artist in our pre-processed cache
	var foundBand *BandInfo
	for _, pa := range processedArtists {
		if pa.Artist.ID == id {
			foundBand = &pa.BandInfo
			break
		}
	}

	// If the artist doesn't exist, return a 404
	if foundBand == nil {
		renderError(w, http.StatusNotFound, "Not Found", "Artist not found.")
		return
	}

	// Load and parse the single artist details template
	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Could not load artist template.")
		return
	}

	// Populate data with the pre-formatted BandInfo structure
	data := struct {
		Title  string
		Artist *BandInfo
	}{
		Title:  foundBand.Name + " - Groupie Tracker",
		Artist: foundBand,
	}

	// Execute and render the artist page
	if err := tmpl.Execute(w, data); err != nil {
		renderError(w, http.StatusInternalServerError, "Internal Server Error", "Template execution failed.")
	}
}

// searchHandler handles asynchronous requests for filtering and searching.
// It returns a JSON array of artists matching all selected criteria for the frontend live search.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse standard filters (year ranges, members, locations)
	filters := ParseFilters(r)

	// Grab the search query for the artist name and normalize it to lowercase
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	var filteredArtists []Artist

	// Loop over pre-processed cache to filter out matching artists
	for _, pa := range processedArtists {
		artist := pa.Artist

		// Apply name search if a query was provided
		if query != "" && !strings.Contains(strings.ToLower(artist.Name), query) {
			continue
		}

		// 1. Filter by Creation Date (numeric comparison)
		if artist.CreationDate < filters.CreationMin || artist.CreationDate > filters.CreationMax {
			continue
		}

		// 2. Filter by First Album Year (uses pre-parsed FirstAlbumYear integer)
		if pa.FirstAlbumYear < filters.FirstAlbumMin || pa.FirstAlbumYear > filters.FirstAlbumMax {
			continue
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
		if filters.Location != "" {
			matchedLocation := false
			for _, locClean := range pa.CleanLocations {
				if strings.Contains(locClean, filters.Location) {
					matchedLocation = true
					break
				}
			}
			if !matchedLocation {
				continue
			}
		}

		// Append the raw Artist struct to the search results
		filteredArtists = append(filteredArtists, artist)
	}

	// Ensure an empty slice is returned as [] instead of null in JSON
	if filteredArtists == nil {
		filteredArtists = []Artist{}
	}

	// Return the filtered artists as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredArtists); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// suggestionsHandler handles asynchronous requests for suggestions
func suggestionsHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SearchSuggestion{})
		return
	}

	var suggestions []SearchSuggestion
	for _, pa := range processedArtists {
		// 1. Match by Artist/Band Name
		if strings.Contains(strings.ToLower(pa.Artist.Name), query) {
			suggestions = append(suggestions, SearchSuggestion{
				DisplayText: pa.Artist.Name + " - artist/band",
				ArtistID:    pa.Artist.ID,
			})
		}

		// 2. Match by Members
		for _, member := range pa.Artist.Members {
			if strings.Contains(strings.ToLower(member), query) {
				suggestions = append(suggestions, SearchSuggestion{
					DisplayText: member + " - member (" + pa.Artist.Name + ")",
					ArtistID:    pa.Artist.ID,
				})
			}
		}

		// 3. Match by Locations
		for _, loc := range pa.BandInfo.Locations {
			if strings.Contains(strings.ToLower(loc), query) {
				suggestions = append(suggestions, SearchSuggestion{
					DisplayText: loc + " - location (" + pa.Artist.Name + ")",
					ArtistID:    pa.Artist.ID,
				})
			}
		}

		// 4. Match by First Album Date
		if strings.Contains(strings.ToLower(pa.Artist.FirstAlbum), query) {
			suggestions = append(suggestions, SearchSuggestion{
				DisplayText: pa.Artist.FirstAlbum + " - first album date (" + pa.Artist.Name + ")",
				ArtistID:    pa.Artist.ID,
			})
		}

		// 5. Match by Creation Date
		creationDateStr := strconv.Itoa(pa.Artist.CreationDate)
		if strings.Contains(creationDateStr, query) {
			suggestions = append(suggestions, SearchSuggestion{
				DisplayText: creationDateStr + " - creation date (" + pa.Artist.Name + ")",
				ArtistID:    pa.Artist.ID,
			})
		}
	}

	// Avoid returning null in JSON by ensuring it's an empty slice if nil
	if suggestions == nil {
		suggestions = []SearchSuggestion{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}


