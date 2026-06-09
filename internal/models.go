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

// LocationsIndex contains all the Locations structs
type LocationsIndex struct {
	Index []Locations `json:"index"`
}

type Locations struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// DatesIndex contains all the Dates structs
type DatesIndex struct {
	Index []Dates `json:"index"`
}

type Dates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// RelationsIndex contains all the Relations structs
type RelationsIndex struct {
	Index []Relations `json:"index"`
}

type Relations struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type BandInfo struct {
	Name         string
	CreationDate int
	Image        string
	Locations    []string
	Dates        []string
	Relations    map[string][]string
	Members      []string
	FirstAlbum   string
}
