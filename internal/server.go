package internal

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Package-level variables to hold cached API data
var (
	cachedArtists      []Artist
	cachedLocationMap  map[int]Locations
	cachedDatesMap     map[int]Dates
	cachedRelationsMap map[int]Relations
	processedArtists   []ProcessedArtist
)

// StartServer initializes the HTTP multiplexer, registers the routes,
// and starts the web server on port 8080.
func StartServer() {
	log.Println("Fetching API data. This might take a moment...")

	var err error
	// 1. Fetch raw Artists list from the external API
	cachedArtists, err = FetchArtists()
	if err != nil {
		log.Fatalf("Failed to fetch artists: %v", err)
	}

	// 2. Fetch raw locations index and build cachedLocationMap keyed by Artist ID
	// for instant lookup when processing relationship objects or filtering.
	locations, err := FetchLocations()
	if err != nil {
		log.Fatalf("Failed to fetch locations: %v", err)
	}
	cachedLocationMap = make(map[int]Locations)
	for _, loc := range locations.Index {
		cachedLocationMap[loc.ID] = loc
	}

	// 3. Fetch raw concert dates index and build cachedDatesMap keyed by Artist ID.
	dates, err := FetchDates()
	if err != nil {
		log.Fatalf("Failed to fetch dates: %v", err)
	}
	cachedDatesMap = make(map[int]Dates)
	for _, d := range dates.Index {
		cachedDatesMap[d.ID] = d
	}

	// 4. Fetch raw relations index mapping locations to dates, keyed by Artist ID.
	relations, err := FetchRelations()
	if err != nil {
		log.Fatalf("Failed to fetch relations: %v", err)
	}
	cachedRelationsMap = make(map[int]Relations)
	for _, r := range relations.Index {
		cachedRelationsMap[r.ID] = r
	}

	// Pre-process and cache artist data to resolve Bottleneck #2
	// This runs exactly once on server startup, shifting computational overhead from request-time to build-time.
	processedArtists = make([]ProcessedArtist, len(cachedArtists))
	for i, artist := range cachedArtists {
		// 1. Pre-parse the First Album year.
		// The format is "DD-MM-YYYY", so we grab the third part (index 2) representing the year.
		var albumYear int
		albumParts := strings.Split(artist.FirstAlbum, "-")
		if len(albumParts) == 3 {
			if yr, err := strconv.Atoi(albumParts[2]); err == nil {
				albumYear = yr
			}
		}

		// 2. Pre-clean locations for search and pre-format locations for templates.
		// - Clean locations: Lowercased and stripped of underscores/hyphens for instant query matching.
		// - Formatted locations: Human-readable strings where underscores and hyphens are replaced with spaces.
		var cleanLocs []string
		var formattedLocations []string
		if locs, exists := cachedLocationMap[artist.ID]; exists {
			for _, loc := range locs.Locations {
				locClean := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(loc, "_", " "), "-", " "))
				cleanLocs = append(cleanLocs, locClean)

				cleanLoc := strings.ReplaceAll(loc, "_", " ")
				cleanLoc = strings.ReplaceAll(cleanLoc, "-", " ")
				formattedLocations = append(formattedLocations, cleanLoc)
			}
		}

		// 3. Pre-format relations (locations mapped to dates) for templates.
		// Underscores and hyphens in location names are cleaned for a better display layout.
		formattedRelations := make(map[string][]string)
		if artistRels, exists := cachedRelationsMap[artist.ID]; exists {
			for loc, datesList := range artistRels.DatesLocations {
				cleanRelLoc := strings.ReplaceAll(loc, "_", " ")
				cleanRelLoc = strings.ReplaceAll(cleanRelLoc, "-", " ")
				formattedRelations[cleanRelLoc] = datesList
			}
		}

		// 4. Pre-format dates for templates.
		// Asterisks (*) are stripped from dates to present clean strings.
		var formattedDates []string
		if artistDates, exists := cachedDatesMap[artist.ID]; exists {
			for _, date := range artistDates.Dates {
				cleanDate := strings.ReplaceAll(date, "*", "")
				formattedDates = append(formattedDates, cleanDate)
			}
		}

		// 5. Build the fully pre-formatted BandInfo struct.
		// This struct will be passed directly to HTML templates during requests,
		// avoiding any memory allocations or string operations when servicing pages.
		bandInfo := BandInfo{
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

		// Save the processed data into our global slice.
		processedArtists[i] = ProcessedArtist{
			Artist:         artist,
			BandInfo:       bandInfo,
			FirstAlbumYear: albumYear,
			CleanLocations: cleanLocs,
		}
	}

	log.Println("API data successfully cached! Starting up the server...")

	// Define the port the server will listen on
	port := ":8080"
	// Create a new HTTP multiplexer (router)
	handler := http.NewServeMux()

	// Print a message to the console indicating the server is starting
	log.Printf("Server is running on port: http://localhost%s\n", port)

	// Register the homeHandler function to handle requests to the root URL ("/")
	handler.HandleFunc("/", homeHandler)
	// Register the artistHandler to display individual artists
	handler.HandleFunc("/artist", artistHandler)
	// Register the search API for client-side asynchronous filtering
	handler.HandleFunc("/api/search", searchHandler)
	// Handle the css call
	fileserver := http.FileServer(http.Dir("./static"))
	handler.Handle("/static/", http.StripPrefix("/static", fileserver))

	// Start the HTTP server; log.Fatal will catch and log any errors if the server fails to start
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}
