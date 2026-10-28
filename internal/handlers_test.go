package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestMain sets up the working directory and populates package-level caches
// with sample test data before any tests run. This ensures that template paths
// resolve correctly and that handlers have data to operate on.
func TestMain(m *testing.M) {
	// Change to the project root so that template paths like "templates/error.html" resolve.
	if err := os.Chdir(".."); err != nil {
		panic("failed to chdir to project root: " + err.Error())
	}

	// --- Sample Artists ---
	cachedArtists = []Artist{
		{
			ID: 1, Name: "Queen", Image: "https://example.com/queen.jpg",
			Members:      []string{"Freddie Mercury", "Brian May", "Roger Taylor", "John Deacon"},
			CreationDate: 1970, FirstAlbum: "14-12-1973",
		},
		{
			ID: 2, Name: "ACDC", Image: "https://example.com/acdc.jpg",
			Members:      []string{"Angus Young", "Malcolm Young", "Brian Johnson"},
			CreationDate: 1973, FirstAlbum: "22-02-1975",
		},
		{
			ID: 3, Name: "Gorillaz", Image: "https://example.com/gorillaz.jpg",
			Members:      []string{"Damon Albarn", "Jamie Hewlett"},
			CreationDate: 1998, FirstAlbum: "26-03-2001",
		},
	}

	// --- Cached Locations ---
	cachedLocationMap = map[int]Locations{
		1: {ID: 1, Locations: []string{"london-uk", "paris-france"}, Dates: ""},
		2: {ID: 2, Locations: []string{"sydney-australia", "new_york-usa"}, Dates: ""},
		3: {ID: 3, Locations: []string{"tokyo-japan", "berlin-germany"}, Dates: ""},
	}

	// --- Cached Dates ---
	cachedDatesMap = map[int]Dates{
		1: {ID: 1, Dates: []string{"*01-07-2020", "*15-08-2020"}},
		2: {ID: 2, Dates: []string{"*10-11-2019", "*25-12-2019"}},
		3: {ID: 3, Dates: []string{"*05-06-2021"}},
	}

	// --- Cached Relations ---
	cachedRelationsMap = map[int]Relations{
		1: {ID: 1, DatesLocations: map[string][]string{
			"london-uk":    {"01-07-2020"},
			"paris-france": {"15-08-2020"},
		}},
		2: {ID: 2, DatesLocations: map[string][]string{
			"sydney-australia": {"10-11-2019"},
			"new_york-usa":     {"25-12-2019"},
		}},
		3: {ID: 3, DatesLocations: map[string][]string{
			"tokyo-japan": {"05-06-2021"},
		}},
	}

	// --- Pre-processed Artists (mirrors the logic in server.go StartServer) ---
	processedArtists = []ProcessedArtist{
		{
			Artist:         cachedArtists[0],
			FirstAlbumYear: 1973,
			CleanLocations: []string{"london uk", "paris france"},
			BandInfo: BandInfo{
				ID: 1, Name: "Queen", CreationDate: 1970,
				Image:      "https://example.com/queen.jpg",
				Locations:  []string{"london uk", "paris france"},
				Dates:      []string{"01-07-2020", "15-08-2020"},
				Relations:  map[string][]string{"london uk": {"01-07-2020"}, "paris france": {"15-08-2020"}},
				Members:    []string{"Freddie Mercury", "Brian May", "Roger Taylor", "John Deacon"},
				FirstAlbum: "14-12-1973",
			},
		},
		{
			Artist:         cachedArtists[1],
			FirstAlbumYear: 1975,
			CleanLocations: []string{"sydney australia", "new york usa"},
			BandInfo: BandInfo{
				ID: 2, Name: "ACDC", CreationDate: 1973,
				Image:      "https://example.com/acdc.jpg",
				Locations:  []string{"sydney australia", "new york usa"},
				Dates:      []string{"10-11-2019", "25-12-2019"},
				Relations:  map[string][]string{"sydney australia": {"10-11-2019"}, "new york usa": {"25-12-2019"}},
				Members:    []string{"Angus Young", "Malcolm Young", "Brian Johnson"},
				FirstAlbum: "22-02-1975",
			},
		},
		{
			Artist:         cachedArtists[2],
			FirstAlbumYear: 2001,
			CleanLocations: []string{"tokyo japan", "berlin germany"},
			BandInfo: BandInfo{
				ID: 3, Name: "Gorillaz", CreationDate: 1998,
				Image:      "https://example.com/gorillaz.jpg",
				Locations:  []string{"tokyo japan", "berlin germany"},
				Dates:      []string{"05-06-2021"},
				Relations:  map[string][]string{"tokyo japan": {"05-06-2021"}},
				Members:    []string{"Damon Albarn", "Jamie Hewlett"},
				FirstAlbum: "26-03-2001",
			},
		},
	}

	os.Exit(m.Run())
}

// ============================================================================
// ParseFilters Tests
// ============================================================================

// TestParseFilters_Defaults verifies that when no query parameters are provided,
// the filter defaults are returned (CreationMin=0, CreationMax=9999, etc.).
func TestParseFilters_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	filters := ParseFilters(req)

	if filters.CreationMin != 0 {
		t.Errorf("expected CreationMin=0, got %d", filters.CreationMin)
	}
	if filters.CreationMax != 9999 {
		t.Errorf("expected CreationMax=9999, got %d", filters.CreationMax)
	}
	if filters.FirstAlbumMin != 0 {
		t.Errorf("expected FirstAlbumMin=0, got %d", filters.FirstAlbumMin)
	}
	if filters.FirstAlbumMax != 9999 {
		t.Errorf("expected FirstAlbumMax=9999, got %d", filters.FirstAlbumMax)
	}
	if len(filters.Members) != 0 {
		t.Errorf("expected empty Members slice, got %v", filters.Members)
	}
	if filters.Location != "" {
		t.Errorf("expected empty Location, got %q", filters.Location)
	}
}

// TestParseFilters_AllParams verifies that all query parameters are correctly
// parsed and assigned to the FilterParams struct.
func TestParseFilters_AllParams(t *testing.T) {
	url := "/?creationDateMin=1960&creationDateMax=2000&firstAlbumMin=1970&firstAlbumMax=1990&members=3&members=4&location=london"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	filters := ParseFilters(req)

	if filters.CreationMin != 1960 {
		t.Errorf("expected CreationMin=1960, got %d", filters.CreationMin)
	}
	if filters.CreationMax != 2000 {
		t.Errorf("expected CreationMax=2000, got %d", filters.CreationMax)
	}
	if filters.FirstAlbumMin != 1970 {
		t.Errorf("expected FirstAlbumMin=1970, got %d", filters.FirstAlbumMin)
	}
	if filters.FirstAlbumMax != 1990 {
		t.Errorf("expected FirstAlbumMax=1990, got %d", filters.FirstAlbumMax)
	}
	if len(filters.Members) != 2 || filters.Members[0] != "3" || filters.Members[1] != "4" {
		t.Errorf("expected Members=[3,4], got %v", filters.Members)
	}
	if filters.Location != "london" {
		t.Errorf("expected Location='london', got %q", filters.Location)
	}
}

// TestParseFilters_InvalidParams verifies that non-numeric values for numeric
// filter parameters gracefully fall back to the defaults.
func TestParseFilters_InvalidParams(t *testing.T) {
	url := "/?creationDateMin=abc&creationDateMax=xyz&firstAlbumMin=!&firstAlbumMax=???"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	filters := ParseFilters(req)

	// Non-numeric strings should leave defaults unchanged
	if filters.CreationMin != 0 {
		t.Errorf("expected CreationMin=0 for invalid input, got %d", filters.CreationMin)
	}
	if filters.CreationMax != 9999 {
		t.Errorf("expected CreationMax=9999 for invalid input, got %d", filters.CreationMax)
	}
	if filters.FirstAlbumMin != 0 {
		t.Errorf("expected FirstAlbumMin=0 for invalid input, got %d", filters.FirstAlbumMin)
	}
	if filters.FirstAlbumMax != 9999 {
		t.Errorf("expected FirstAlbumMax=9999 for invalid input, got %d", filters.FirstAlbumMax)
	}
}

// TestParseFilters_LocationCommaStrip verifies that commas are stripped from
// the location parameter (e.g. "London, UK" → "london uk").
func TestParseFilters_LocationCommaStrip(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?location=London%2C+UK", nil)
	filters := ParseFilters(req)

	expected := "london uk"
	if filters.Location != expected {
		t.Errorf("expected Location=%q, got %q", expected, filters.Location)
	}
}

// TestParseFilters_MultipleMembers verifies that multiple "members" query
// parameters are correctly captured as a slice.
func TestParseFilters_MultipleMembers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?members=1&members=2&members=5", nil)
	filters := ParseFilters(req)

	if len(filters.Members) != 3 {
		t.Fatalf("expected 3 members params, got %d", len(filters.Members))
	}
	expected := []string{"1", "2", "5"}
	for i, v := range expected {
		if filters.Members[i] != v {
			t.Errorf("expected Members[%d]=%q, got %q", i, v, filters.Members[i])
		}
	}
}

// ============================================================================
// homeHandler Tests
// ============================================================================

// TestHomeHandler_RootPath verifies that a GET request to "/" returns HTTP 200.
func TestHomeHandler_RootPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	homeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for root path, got %d", rr.Code)
	}
}

// TestHomeHandler_NotFoundPath verifies that a GET request to a non-root path
// returns HTTP 404 via renderError.
func TestHomeHandler_NotFoundPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()

	homeHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for non-root path, got %d", rr.Code)
	}
}

// ============================================================================
// artistHandler Tests
// ============================================================================

// TestArtistHandler_MissingID verifies that omitting the "id" query parameter
// results in a 400 Bad Request.
func TestArtistHandler_MissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist", nil)
	rr := httptest.NewRecorder()

	artistHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing ID, got %d", rr.Code)
	}
}

// TestArtistHandler_InvalidID verifies that a non-numeric "id" parameter
// results in a 400 Bad Request.
func TestArtistHandler_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=abc", nil)
	rr := httptest.NewRecorder()

	artistHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid ID, got %d", rr.Code)
	}
}

// TestArtistHandler_NegativeID verifies that a negative "id" parameter
// results in a 400 Bad Request (id < 1 is rejected).
func TestArtistHandler_NegativeID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=-1", nil)
	rr := httptest.NewRecorder()

	artistHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for negative ID, got %d", rr.Code)
	}
}

// TestArtistHandler_NotFound verifies that a valid but non-existent artist ID
// returns a 404 Not Found.
func TestArtistHandler_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=9999", nil)
	rr := httptest.NewRecorder()

	artistHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for non-existent artist, got %d", rr.Code)
	}
}

// TestArtistHandler_ValidID verifies that a valid existing artist ID returns
// HTTP 200 with the artist page rendered.
func TestArtistHandler_ValidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artist?id=1", nil)
	rr := httptest.NewRecorder()

	artistHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for valid artist, got %d", rr.Code)
	}

	// The response body should contain the artist name
	body := rr.Body.String()
	if !strings.Contains(body, "Queen") {
		t.Error("expected response body to contain 'Queen'")
	}
}

// ============================================================================
// searchHandler Tests
// ============================================================================

// TestSearchHandler_EmptyResults verifies that a search query with no matches
// returns an empty JSON array (not null).
func TestSearchHandler_EmptyResults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=zzzznonexistent", nil)
	rr := httptest.NewRecorder()

	searchHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var results []Artist
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(results))
	}
}

// TestSearchHandler_WithResults verifies that a matching search query returns
// a non-empty JSON array of artists.
func TestSearchHandler_WithResults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=queen", nil)
	rr := httptest.NewRecorder()

	searchHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var results []Artist
	if err := json.NewDecoder(rr.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least 1 result for 'queen' query, got 0")
	}
	// Verify the matched artist is indeed Queen
	if results[0].Name != "Queen" {
		t.Errorf("expected first result to be 'Queen', got %q", results[0].Name)
	}
}

// TestSearchHandler_ContentType verifies that the search handler returns
// the correct Content-Type header.
func TestSearchHandler_ContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
	rr := httptest.NewRecorder()

	searchHandler(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

// ============================================================================
// suggestionsHandler Tests
// ============================================================================

// TestSuggestionsHandler_EmptyQuery verifies that an empty "q" parameter
// returns an empty JSON array of suggestions.
func TestSuggestionsHandler_EmptyQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/suggestions?q=", nil)
	rr := httptest.NewRecorder()

	suggestionsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var suggestions []SearchSuggestion
	if err := json.NewDecoder(rr.Body).Decode(&suggestions); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for empty query, got %d", len(suggestions))
	}
}

// TestSuggestionsHandler_WithResults verifies that a matching query returns
// relevant search suggestions.
func TestSuggestionsHandler_WithResults(t *testing.T) {
	// "freddie" should match the member "Freddie Mercury" of Queen
	req := httptest.NewRequest(http.MethodGet, "/api/suggestions?q=freddie", nil)
	rr := httptest.NewRecorder()

	suggestionsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var suggestions []SearchSuggestion
	if err := json.NewDecoder(rr.Body).Decode(&suggestions); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if len(suggestions) == 0 {
		t.Fatal("expected at least 1 suggestion for 'freddie', got 0")
	}

	// Verify the suggestion refers to the correct member and artist
	found := false
	for _, s := range suggestions {
		if s.Value == "Freddie Mercury" && s.Type == "member" && s.Info == "Queen" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a suggestion for 'Freddie Mercury' (member of Queen), got %+v", suggestions)
	}
}

// TestSuggestionsHandler_ContentType verifies that the suggestions handler
// returns the correct Content-Type header.
func TestSuggestionsHandler_ContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/suggestions?q=test", nil)
	rr := httptest.NewRecorder()

	suggestionsHandler(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

// ============================================================================
// renderError Tests
// ============================================================================

// TestRenderError_404 verifies that renderError produces a 404 status code
// and renders the error template with the correct content.
func TestRenderError_404(t *testing.T) {
	rr := httptest.NewRecorder()

	renderError(rr, http.StatusNotFound, "Not Found", "The page does not exist.")

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	body := rr.Body.String()
	// The error template should include the status code, title, and message
	if !strings.Contains(body, "404") {
		t.Error("expected response body to contain '404'")
	}
	if !strings.Contains(body, "Not Found") {
		t.Error("expected response body to contain 'Not Found'")
	}
	if !strings.Contains(body, "The page does not exist.") {
		t.Error("expected response body to contain the error message")
	}
}

// TestRenderError_500 verifies that renderError produces a 500 status code
// and renders the error template with the correct content.
func TestRenderError_500(t *testing.T) {
	rr := httptest.NewRecorder()

	renderError(rr, http.StatusInternalServerError, "Internal Server Error", "Something broke.")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "500") {
		t.Error("expected response body to contain '500'")
	}
	if !strings.Contains(body, "Internal Server Error") {
		t.Error("expected response body to contain 'Internal Server Error'")
	}
	if !strings.Contains(body, "Something broke.") {
		t.Error("expected response body to contain the error message")
	}
}
