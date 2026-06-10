package internal

import (
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

	// Fetch the artist data from the external API
	artists, err := FetchArtists()
	// Handle any errors that occur during the API fetch
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch artists.")
		return
	}

	locations, err := FetchLocations()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch locations.")
		return
	}

	dates, err := FetchDates()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch dates.")
		return
	}

	relations, err := FetchRelations()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch relations.")
		return
	}

	// Create lookup maps by ID for robust data matching
	locationMap := make(map[int]Locations)
	for _, loc := range locations.Index {
		locationMap[loc.ID] = loc
	}

	datesMap := make(map[int]Dates)
	for _, d := range dates.Index {
		datesMap[d.ID] = d
	}

	relationsMap := make(map[int]Relations)
	for _, r := range relations.Index {
		relationsMap[r.ID] = r
	}

	filters := ParseFilters(r)

	var bands []BandInfo

	for _, artist := range artists {
		// 1. creation date filter
		if artist.CreationDate < filters.CreationMin || artist.CreationDate > filters.CreationMax {
			continue
		}
		// 2. first album filter
		albumParts := strings.Split(artist.FirstAlbum, "-")
		if len(albumParts) == 3 {
			albumYear, err := strconv.Atoi(albumParts[2])
			if err == nil {
				if albumYear < filters.FirstAlbumMin || albumYear > filters.FirstAlbumMax {
					continue
				}
			}
		}
		// 3. filter members
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

		// 4. filter location
		artistLocs, locExists := locationMap[artist.ID]
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
		artistRels, relExists := relationsMap[artist.ID]
		if relExists {
			for loc, datesList := range artistRels.DatesLocations {
				cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
				cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
				formattedRelations[cleanRelLoc] = datesList
			}
		}

		var formattedDates []string
		artistDates, dateExists := datesMap[artist.ID]
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

	artists, err := FetchArtists()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch artists.")
		return
	}
	relations, err := FetchRelations()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "API Error", "Failed to fetch relations.")
		return
	}

	var foundBand *BandInfo
	for _, artist := range artists {
		if artist.ID == id {
			formattedRelations := make(map[string][]string)

			// Match relation by ID safely
			for _, rel := range relations.Index {
				if rel.ID == artist.ID {
					for loc, datesList := range rel.DatesLocations {
						cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
						cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
						formattedRelations[cleanRelLoc] = datesList
					}
					break
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
