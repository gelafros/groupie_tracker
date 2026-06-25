package internal

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
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

	// Grab the search query and scope and normalize to lowercase
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	var filteredArtists []Artist

	// Loop over pre-processed cache to filter out matching artists
	for _, pa := range processedArtists {
		artist := pa.Artist

		// Apply search logic for the 5 search cases if query is provided
		if query != "" {
			matched := false

			// 1. artist/band name
			if scope == "" || scope == "all" || scope == "artist/band" {
				if strings.Contains(strings.ToLower(artist.Name), query) {
					matched = true
				}
			}

			// 2. members
			if !matched && (scope == "" || scope == "all" || scope == "member") {
				for _, m := range artist.Members {
					if strings.Contains(strings.ToLower(m), query) {
						matched = true
						break
					}
				}
			}

			// 3. locations
			if !matched && (scope == "" || scope == "all" || scope == "location") {
				if locs, exists := cachedLocationMap[artist.ID]; exists {
					for _, loc := range locs.Locations {
						locLower := strings.ToLower(loc)
						// Compare with raw location name (with hyphens/underscores) and clean version (with spaces)
						locClean := strings.ReplaceAll(strings.ReplaceAll(locLower, "_", " "), "-", " ")
						queryClean := strings.ReplaceAll(strings.ReplaceAll(query, "_", " "), "-", " ")
						if strings.Contains(locLower, query) || strings.Contains(locClean, queryClean) {
							matched = true
							break
						}
					}
				}
			}

			// 4. first album date
			if !matched && (scope == "" || scope == "all" || scope == "first album date") {
				if strings.Contains(strings.ToLower(artist.FirstAlbum), query) {
					matched = true
				}
			}

			// 5. creation date
			if !matched && (scope == "" || scope == "all" || scope == "creation date") {
				creationDateStr := strconv.Itoa(artist.CreationDate)
				if strings.Contains(creationDateStr, query) {
					matched = true
				}
			}

			if !matched {
				continue
			}
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
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SearchSuggestion{})
		return
	}

	type sortedSuggestion struct {
		suggestion SearchSuggestion
		matchIndex int
		matchValue string
	}

	var tempSuggestions []sortedSuggestion
	seen := make(map[string]bool)

	for _, pa := range processedArtists {
		// 1. Match by Artist/Band Name
		if scope == "" || scope == "all" || scope == "artist/band" {
			if strings.Contains(strings.ToLower(pa.Artist.Name), query) {
				text := pa.Artist.Name + " - artist/band"
				if !seen[text] {
					seen[text] = true
					idx := strings.Index(strings.ToLower(pa.Artist.Name), query)
					tempSuggestions = append(tempSuggestions, sortedSuggestion{
						suggestion: SearchSuggestion{
							Value:    pa.Artist.Name,
							Type:     "artist/band",
							Info:     "",
							ArtistID: pa.Artist.ID,
						},
						matchIndex: idx,
						matchValue: pa.Artist.Name,
					})
				}
			}
		}

		// 2. Match by Members
		if scope == "" || scope == "all" || scope == "member" {
			for _, member := range pa.Artist.Members {
				if strings.Contains(strings.ToLower(member), query) {
					text := member + " - member (" + pa.Artist.Name + ")"
					if !seen[text] {
						seen[text] = true
						idx := strings.Index(strings.ToLower(member), query)
						tempSuggestions = append(tempSuggestions, sortedSuggestion{
							suggestion: SearchSuggestion{
								Value:    member,
								Type:     "member",
								Info:     pa.Artist.Name,
								ArtistID: pa.Artist.ID,
							},
							matchIndex: idx,
							matchValue: member,
						})
					}
				}
			}
		}

		// 3. Match by Locations
		if scope == "" || scope == "all" || scope == "location" {
			if locs, exists := cachedLocationMap[pa.Artist.ID]; exists {
				for _, loc := range locs.Locations {
					locLower := strings.ToLower(loc)
					locClean := strings.ReplaceAll(strings.ReplaceAll(locLower, "_", " "), "-", " ")
					queryClean := strings.ReplaceAll(strings.ReplaceAll(query, "_", " "), "-", " ")

					idx := -1
					if strings.Contains(locLower, query) {
						idx = strings.Index(locLower, query)
					} else if strings.Contains(locClean, queryClean) {
						idx = strings.Index(locClean, queryClean)
					}

					if idx != -1 {
						text := loc + " - location (" + pa.Artist.Name + ")"
						if !seen[text] {
							seen[text] = true
							tempSuggestions = append(tempSuggestions, sortedSuggestion{
								suggestion: SearchSuggestion{
									Value:    loc,
									Type:     "location",
									Info:     pa.Artist.Name,
									ArtistID: pa.Artist.ID,
								},
								matchIndex: idx,
								matchValue: loc,
							})
						}
					}
				}
			}
		}

		// 4. Match by First Album Date
		if scope == "" || scope == "all" || scope == "first album date" {
			if strings.Contains(strings.ToLower(pa.Artist.FirstAlbum), query) {
				text := pa.Artist.FirstAlbum + " - first album date (" + pa.Artist.Name + ")"
				if !seen[text] {
					seen[text] = true
					idx := strings.Index(strings.ToLower(pa.Artist.FirstAlbum), query)
					tempSuggestions = append(tempSuggestions, sortedSuggestion{
						suggestion: SearchSuggestion{
							Value:    pa.Artist.FirstAlbum,
							Type:     "first album date",
							Info:     pa.Artist.Name,
							ArtistID: pa.Artist.ID,
						},
						matchIndex: idx,
						matchValue: pa.Artist.FirstAlbum,
					})
				}
			}
		}

		// 5. Match by Creation Date
		if scope == "" || scope == "all" || scope == "creation date" {
			creationDateStr := strconv.Itoa(pa.Artist.CreationDate)
			if strings.Contains(creationDateStr, query) {
				text := creationDateStr + " - creation date (" + pa.Artist.Name + ")"
				if !seen[text] {
					seen[text] = true
					idx := strings.Index(creationDateStr, query)
					tempSuggestions = append(tempSuggestions, sortedSuggestion{
						suggestion: SearchSuggestion{
							Value:    creationDateStr,
							Type:     "creation date",
							Info:     pa.Artist.Name,
							ArtistID: pa.Artist.ID,
						},
						matchIndex: idx,
						matchValue: creationDateStr,
					})
				}
			}
		}
	}

	// Sort suggestions: first by matchIndex (so prefix matches are shown first),
	// second by matched value alphabetically, third by value.
	sort.Slice(tempSuggestions, func(i, j int) bool {
		if tempSuggestions[i].matchIndex != tempSuggestions[j].matchIndex {
			return tempSuggestions[i].matchIndex < tempSuggestions[j].matchIndex
		}
		valI := strings.ToLower(tempSuggestions[i].matchValue)
		valJ := strings.ToLower(tempSuggestions[j].matchValue)
		if valI != valJ {
			return valI < valJ
		}
		return strings.ToLower(tempSuggestions[i].suggestion.Value) < strings.ToLower(tempSuggestions[j].suggestion.Value)
	})

	suggestions := make([]SearchSuggestion, len(tempSuggestions))
	for i, ts := range tempSuggestions {
		suggestions[i] = ts.suggestion
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
