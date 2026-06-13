package internal

// Artist represents the data structure for a single band/artist from the external API.
type Artist struct {
	ID           int      `json:"id"`           // Unique identifier for the artist/band.
	Image        string   `json:"image"`        // URL pointing to the artist's thumbnail image.
	Name         string   `json:"name"`         // Name of the band or solo artist.
	Members      []string `json:"members"`      // Names of the members belonging to the band.
	CreationDate int      `json:"creationDate"` // The year when the band/artist was formed.
	FirstAlbum   string   `json:"firstAlbum"`   // Release date of their first album (typically formatted as "DD-MM-YYYY").
	Locations    string   `json:"locations"`    // API URL endpoint containing concert location details.
	ConcertDates string   `json:"concertDates"` // API URL endpoint containing concert date details.
	Relations    string   `json:"relations"`    // API URL endpoint mapping locations to their concert dates.
}

// LocationsIndex represents the root JSON response from the locations API endpoint,
// which contains a wrapper index containing a list of Locations.
type LocationsIndex struct {
	Index []Locations `json:"index"` // List of concert locations grouped by artist.
}

// Locations represents the raw concert locations structure for a specific artist.
type Locations struct {
	ID        int      `json:"id"`        // Matches the corresponding Artist ID.
	Locations []string `json:"locations"` // List of raw location strings (e.g. "london-uk", "paris-france").
	Dates     string   `json:"dates"`     // API URL endpoint pointing to the exact concert dates.
}

// DatesIndex represents the root JSON response from the dates API endpoint,
// which contains a wrapper index containing a list of Dates.
type DatesIndex struct {
	Index []Dates `json:"index"` // List of concert dates grouped by artist.
}

// Dates represents the raw concert dates structure for a specific artist.
type Dates struct {
	ID    int      `json:"id"`    // Matches the corresponding Artist ID.
	Dates []string `json:"dates"` // List of raw concert date strings (prefixed with asterisks in raw JSON).
}

// RelationsIndex represents the root JSON response from the relations API endpoint,
// which contains a wrapper index containing a list of Relations.
type RelationsIndex struct {
	Index []Relations `json:"index"` // List of relation structures mapping locations to dates.
}

// Relations links a specific artist's concert locations to their corresponding dates.
// This is the direct mapping of where and when a band played.
type Relations struct {
	ID             int                 `json:"id"`             // Matches the corresponding Artist ID.
	DatesLocations map[string][]string `json:"datesLocations"` // Map of raw location names (keys) to list of date strings (values).
}

// BandInfo is a consolidated, clean struct that holds all merged and formatted data
// for a specific artist. This is passed directly to HTML templates.
type BandInfo struct {
	ID           int                 // Unique ID of the artist.
	Name         string              // Name of the artist/band.
	CreationDate int                 // Year the band was formed.
	Image        string              // URL path to the thumbnail image.
	Locations    []string            // Cleaned list of concert locations (e.g., "London Uk").
	Dates        []string            // Cleaned list of concert dates (no leading asterisks).
	Relations    map[string][]string // Cleaned location-to-date schedule map.
	Members      []string            // List of member names.
	FirstAlbum   string              // First album release date string.
}

// FilterParams holds the parsed search and filtering criteria submitted by the user.
type FilterParams struct {
	CreationMin   int      // Minimum creation year boundary.
	CreationMax   int      // Maximum creation year boundary.
	FirstAlbumMin int      // Minimum first album release year boundary.
	FirstAlbumMax int      // Maximum first album release year boundary.
	Members       []string // Slice of member counts requested (e.g. ["1", "2", "4"]).
	Location      string   // Search term matching concert locations.
}

// ProcessedArtist holds the raw artist data along with pre-parsed fields for fast filtering.
// This is the optimized in-memory model that avoids doing expensive string transformations
// and integer conversions during client HTTP requests.
type ProcessedArtist struct {
	Artist Artist // The raw, unmutated artist struct directly from the API.

	BandInfo       BandInfo // The pre-formatted artist details ready for rendering in templates.
	FirstAlbumYear int      // The pre-parsed year of the first album as an integer (e.g. 1979) for range filtering.
	CleanLocations []string // Lowercased concert locations with underscores and hyphens replaced with spaces.
}
