package internal

import (
	"net/http"
	"strconv"
	"strings"
)

// ParseFilters extracts and parses the filtering criteria from the incoming HTTP request.
// It populates a FilterParams struct with default fallback boundaries if queries are missing.
func ParseFilters(r *http.Request) FilterParams {
	// Retrieve the raw query parameters map from the HTTP request URL
	queryParams := r.URL.Query()
	creationMinStr := queryParams.Get("creationDateMin")
	creationMaxStr := queryParams.Get("creationDateMax")
	firstAlbumMinStr := queryParams.Get("firstAlbumMin")
	firstAlbumMaxStr := queryParams.Get("firstAlbumMax")
	membersParams := queryParams["members"] // Grabs all array parameters matching ?members=X&members=Y...

	// Normalize the location search input: lowercase it, trim surrounding spaces, and remove commas
	// to make matching more flexible (e.g. matching "London, UK" with "london uk").
	locationParam := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(queryParams.Get("location"))), ",", "")

	// Initialize the filters structure with wide defaults (0 to 9999) so that if the user
	// hasn't specified range limits, all artists are matched by default.
	filters := FilterParams{
		CreationMin:   0,
		CreationMax:   9999,
		FirstAlbumMin: 0,
		FirstAlbumMax: 9999,
		Members:       membersParams,
		Location:      locationParam,
	}

	// 1. Convert Creation Date Minimum from string to integer if provided
	if creationMinStr != "" {
		if val, err := strconv.Atoi(creationMinStr); err == nil {
			filters.CreationMin = val
		}
	}

	// 2. Convert Creation Date Maximum from string to integer if provided
	if creationMaxStr != "" {
		if val, err := strconv.Atoi(creationMaxStr); err == nil {
			filters.CreationMax = val
		}
	}

	// 3. Convert First Album Minimum Year from string to integer if provided
	if firstAlbumMinStr != "" {
		if val, err := strconv.Atoi(firstAlbumMinStr); err == nil {
			filters.FirstAlbumMin = val
		}
	}

	// 4. Convert First Album Maximum Year from string to integer if provided
	if firstAlbumMaxStr != "" {
		if val, err := strconv.Atoi(firstAlbumMaxStr); err == nil {
			filters.FirstAlbumMax = val
		}
	}

	return filters
}
