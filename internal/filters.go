package internal

import (
	"net/http"
	"strconv"
	"strings"
)

// ParseFilters extracts filter criteria from the request query parameters.
// It populates a FilterParams struct, using default values if parameters are missing.
func ParseFilters(r *http.Request) FilterParams {
	// Retrieve query parameters from the HTTP request
	queryParams := r.URL.Query()
	creationMinStr := queryParams.Get("creationDateMin")
	creationMaxStr := queryParams.Get("creationDateMax")
	firstAlbumMinStr := queryParams.Get("firstAlbumMin")
	firstAlbumMaxStr := queryParams.Get("firstAlbumMax")
	membersParams := queryParams["members"]
	locationParam := strings.ToLower(strings.TrimSpace(queryParams.Get("location")))

	// Initialize the filters struct with wide default ranges to include all artists by default
	filters := FilterParams{
		CreationMin:   0,
		CreationMax:   9999,
		FirstAlbumMin: 0,
		FirstAlbumMax: 9999,
		Members:       membersParams,
		Location:      locationParam,
	}

	// Convert string parameters to integers and update the struct if valid
	if creationMinStr != "" {
		if val, err := strconv.Atoi(creationMinStr); err == nil {
			filters.CreationMin = val
		}
	}
	// Parse creation max date
	if creationMaxStr != "" {
		if val, err := strconv.Atoi(creationMaxStr); err == nil {
			filters.CreationMax = val
		}
	}
	// Parse first album min date
	if firstAlbumMinStr != "" {
		if val, err := strconv.Atoi(firstAlbumMinStr); err == nil {
			filters.FirstAlbumMin = val
		}
	}
	// Parse first album max date
	if firstAlbumMaxStr != "" {
		if val, err := strconv.Atoi(firstAlbumMaxStr); err == nil {
			filters.FirstAlbumMax = val
		}
	}

	return filters
}
