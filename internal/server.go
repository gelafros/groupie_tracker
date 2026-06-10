package internal

import (
	"log"
	"net/http"
)

// Package-level variables to hold cached API data
var (
	cachedArtists      []Artist
	cachedLocationMap  map[int]Locations
	cachedDatesMap     map[int]Dates
	cachedRelationsMap map[int]Relations
)

// StartServer initializes the HTTP multiplexer, registers the routes,
// and starts the web server on port 8080.
func StartServer() {
	log.Println("Fetching API data. This might take a moment...")

	var err error
	cachedArtists, err = FetchArtists()
	if err != nil {
		log.Fatalf("Failed to fetch artists: %v", err)
	}

	locations, err := FetchLocations()
	if err != nil {
		log.Fatalf("Failed to fetch locations: %v", err)
	}
	cachedLocationMap = make(map[int]Locations)
	for _, loc := range locations.Index {
		cachedLocationMap[loc.ID] = loc
	}

	dates, err := FetchDates()
	if err != nil {
		log.Fatalf("Failed to fetch dates: %v", err)
	}
	cachedDatesMap = make(map[int]Dates)
	for _, d := range dates.Index {
		cachedDatesMap[d.ID] = d
	}

	relations, err := FetchRelations()
	if err != nil {
		log.Fatalf("Failed to fetch relations: %v", err)
	}
	cachedRelationsMap = make(map[int]Relations)
	for _, r := range relations.Index {
		cachedRelationsMap[r.ID] = r
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
	// Handle the css call
	fileserver := http.FileServer(http.Dir("./static"))
	handler.Handle("/static/", http.StripPrefix("/static", fileserver))

	// Start the HTTP server; log.Fatal will catch and log any errors if the server fails to start
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}
