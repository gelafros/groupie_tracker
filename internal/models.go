package internal

// Artist represents the data structure for a single band/artist from the API.
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

// LocationsIndex represents the root response from the locations API containing all locations.
type LocationsIndex struct {
	Index []Locations `json:"index"`
}

// Locations represents the concert locations for a specific artist.
type Locations struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// DatesIndex represents the root response from the dates API containing all concert dates.
type DatesIndex struct {
	Index []Dates `json:"index"`
}

// Dates represents the concert dates for a specific artist.
type Dates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// RelationsIndex represents the root response from the relations API containing all relations.
type RelationsIndex struct {
	Index []Relations `json:"index"`
}

// Relations links a specific artist's concert locations to their corresponding dates.
type Relations struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// BandInfo is a consolidated struct that holds all merged data for a specific artist to be sent to templates.
type BandInfo struct {
	ID           int
	Name         string
	CreationDate int
	Image        string
	Locations    []string
	Dates        []string
	Relations    map[string][]string
	Members      []string
	FirstAlbum   string
}

// FilterParams holds all the parsed criteria for filtering artists.
type FilterParams struct {
	CreationMin   int
	CreationMax   int
	FirstAlbumMin int
	FirstAlbumMax int
	Members       []string
	Location      string
}

// ProcessedArtist holds the raw artist data along with pre-parsed fields for fast filtering.
// This is the optimized in-memory model that avoids doing expensive string transformations
// and integer conversions during client HTTP requests.
type ProcessedArtist struct {
	Artist         Artist       // The raw, unmutated artist struct directly from the API.
	BandInfo       BandInfo     // The pre-formatted artist details ready for rendering in templates.
	FirstAlbumYear int          // The pre-parsed year of the first album as an integer (e.g. 1979) for range filtering.
	CleanLocations []string     // Lowercased concert locations with underscores and hyphens replaced with spaces.
}
