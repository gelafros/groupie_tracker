package internal

import (
	"net/http"
	"strconv"
	"strings"
)

func ParseFilters(r *http.Request) FilterParams {
	// get query params
	queryParams := r.URL.Query()
	creationMinStr := queryParams.Get("creationDateMin")
	creationMaxStr := queryParams.Get("creationDateMax")
	firstAlbumMinStr := queryParams.Get("firstAlbumMin")
	firstAlbumMaxStr := queryParams.Get("firstAlbumMax")
	membersParams := queryParams["members"]
	locationParam := strings.ToLower(strings.TrimSpace(queryParams.Get("location")))

	filters := FilterParams{
		CreationMin:   0,
		CreationMax:   9999,
		FirstAlbumMin: 0,
		FirstAlbumMax: 9999,
		Members:       membersParams,
		Location:      locationParam,
	}

	// convert strings to integers
	if creationMinStr != "" {
		if val, err := strconv.Atoi(creationMinStr); err == nil {
			filters.CreationMin = val
		}
	}
	if creationMaxStr != "" {
		if val, err := strconv.Atoi(creationMaxStr); err == nil {
			filters.CreationMax = val
		}
	}
	if firstAlbumMinStr != "" {
		if val, err := strconv.Atoi(firstAlbumMinStr); err == nil {
			filters.FirstAlbumMin = val
		}
	}
	if firstAlbumMaxStr != "" {
		if val, err := strconv.Atoi(firstAlbumMaxStr); err == nil {
			filters.FirstAlbumMax = val
		}
	}

	return filters
}
