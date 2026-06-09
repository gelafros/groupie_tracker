package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchArtists() ([]Artist, error) {
	// Define the API endpoint URL for artists
	url := "https://groupietrackers.herokuapp.com/api/artists"

	// Make an HTTP GET request to the API
	resp, err := http.Get(url)
	// Handle any network or request errors
	if err != nil {
		fmt.Printf("Failed to fetch artists from %s\n", url)
		return nil, err
	}
	// Ensure the response body is closed after the function finishes to prevent memory leaks
	defer resp.Body.Close()

	// Initialize an empty slice to hold the decoded artist data
	var artists []Artist

	// Create a JSON decoder and parse the response body directly into the artists slice
	err = json.NewDecoder(resp.Body).Decode(&artists)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode artists from %s\n", url)
		return artists, err
	}

	// Return the successfully populated list of artists
	return artists, nil
}

func FetchLocations() (LocationsIndex, error) {
	// Define the API endpoint URL for artists
	url := "https://groupietrackers.herokuapp.com/api/locations"

	// Make an HTTP GET request to the API
	resp, err := http.Get(url)
	// Handle any network or request errors
	if err != nil {
		fmt.Printf("Failed to fetch locations from %s\n", url)
		return LocationsIndex{}, err
	}
	// Ensure the response body is closed after the function finishes to prevent memory leaks
	defer resp.Body.Close()

	// Initialize an empty slice to hold the decoded locations data
	var locations LocationsIndex

	// Create a JSON decoder and parse the response body directly into the artists slice
	err = json.NewDecoder(resp.Body).Decode(&locations)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode artists from %s\n", url)
		return locations, err
	}

	// Return the successfully populated list of artists
	return locations, nil
}

func FetchDates() (DatesIndex, error) {
	url := "https://groupietrackers.herokuapp.com/api/dates"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch dates from %s\n", url)
		return DatesIndex{}, err
	}

	defer resp.Body.Close()

	var dates DatesIndex

	err = json.NewDecoder(resp.Body).Decode(&dates)
	if err != nil {
		fmt.Printf("Failed to decode artists from %s\n", url)
		return dates, err
	}

	return dates, nil
}

func FetchRelations() (RelationsIndex, error) {
	// Define the API endpoint URL for artists
	url := "https://groupietrackers.herokuapp.com/api/relation"

	// Make an HTTP GET request to the API
	resp, err := http.Get(url)
	// Handle any network or request errors
	if err != nil {
		fmt.Printf("Failed to fetch relations from %s\n", url)
		return RelationsIndex{}, err
	}
	// Ensure the response body is closed after the function finishes to prevent memory leaks
	defer resp.Body.Close()

	// Initialize an empty slice to hold the decoded artist data
	var relations RelationsIndex

	// Create a JSON decoder and parse the response body directly into the artists slice
	err = json.NewDecoder(resp.Body).Decode(&relations)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode relations from %s\n", url)
		return relations, err
	}

	// Return the successfully populated list of artists
	return relations, nil
}
