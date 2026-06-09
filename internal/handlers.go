package internal

import (
	"html/template"
	"net/http"
	"strings"
)

// homeHandler parses and serves the index.html template for the root route.
// It handles template parsing and execution errors by responding with a 500 status.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure that only the exact root path "/" is handled here to prevent it from acting as a catch-all
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Parse the HTML template file
	output, err := template.ParseFiles("templates/index.html")
	// Handle any errors that occur during parsing (e.g., file not found)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Fetch the artist data from the external API
	artists, err := FetchArtists()
	// Handle any errors that occur during the API fetch
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	locations, err := FetchLocations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dates, err := FetchDates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	relations, err := FetchRelations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bands []BandInfo

	for i, artist := range artists {

		var formattedLocations []string

		for _, location := range locations.Index[i].Locations {
			cleanLoc := strings.ReplaceAll(location, "_", " ")
			cleanLoc = strings.ReplaceAll(cleanLoc, "-", " ")
			formattedLocations = append(formattedLocations, cleanLoc)
		}

		formattedRelations := make(map[string][]string)
		for loc, datesList := range relations.Index[i].DatesLocations {
			cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
			cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
			formattedRelations[cleanRelLoc] = datesList
		}

		var formattedDates []string
		for _, date := range dates.Index[i].Dates {
			cleanDate := strings.ReplaceAll(date, "*", "")
			formattedDates = append(formattedDates, cleanDate)
		}

		band := BandInfo{
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
	err = output.Execute(w, bands)
	// Handle any errors that occur during template execution
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
