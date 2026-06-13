package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// FetchArtists requests and decodes the list of musical artists and bands from the API.
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

// FetchLocations requests and decodes the locations wrapper index from the API.
func FetchLocations() (LocationsIndex, error) {
	// Define the API endpoint URL for locations
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

	// Initialize an empty structure to hold the decoded locations data
	var locations LocationsIndex

	// Create a JSON decoder and parse the response body directly into the locations struct
	err = json.NewDecoder(resp.Body).Decode(&locations)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode locations from %s\n", url)
		return locations, err
	}

	// Return the successfully populated locations struct
	return locations, nil
}

// FetchDates requests and decodes the concert dates wrapper index from the API.
func FetchDates() (DatesIndex, error) {
	// Define the API endpoint URL for concert dates
	url := "https://groupietrackers.herokuapp.com/api/dates"

	// Make an HTTP GET request to the API
	resp, err := http.Get(url)
	// Handle any network or request errors
	if err != nil {
		fmt.Printf("Failed to fetch dates from %s\n", url)
		return DatesIndex{}, err
	}
	// Ensure the response body is closed after the function finishes to prevent memory leaks
	defer resp.Body.Close()

	// Initialize an empty structure to hold the decoded dates data
	var dates DatesIndex

	// Create a JSON decoder and parse the response body directly into the dates struct
	err = json.NewDecoder(resp.Body).Decode(&dates)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode dates from %s\n", url)
		return dates, err
	}

	// Return the successfully populated dates struct
	return dates, nil
}

// FetchRelations requests and decodes the locations-to-dates relation wrapper index from the API.
func FetchRelations() (RelationsIndex, error) {
	// Define the API endpoint URL for artist relations mapping
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

	// Initialize an empty structure to hold the decoded relations data
	var relations RelationsIndex

	// Create a JSON decoder and parse the response body directly into the relations struct
	err = json.NewDecoder(resp.Body).Decode(&relations)
	// Handle any errors that occur during JSON decoding
	if err != nil {
		fmt.Printf("Failed to decode relations from %s\n", url)
		return relations, err
	}

	// Return the successfully populated relations struct
	return relations, nil
}
