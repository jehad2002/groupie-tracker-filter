package main

type AR struct {
	ID         int      `json:"id"`
	Image      string   `json:"image"`
	Name       string   `json:"name"`
	Members    []string `json:"members"`
	Creation   int      `json:"creationDate"`
	FirstAlbum string   `json:"firstAlbum"`
}

type Location struct {
	Locations []string `json:"locations"`
}

type AllLocation struct {
	Index []struct {
		//ID        int      `json:"id"`
		Locationss []string `json:"locations"`
		//Dates     string   `json:"dates"`
	} `json:"index"`
}

type Dates struct {
	Dates []string `json:"dates"`
}

type Relationship struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

type Details struct {
	Artist            AR
	Locations         Location
	ConcertDates      Dates
	DatesAndLocations Relationship
}

type Summary struct {
	ID    int    `json:"id"`
	Image string `json:"image"`
	Name  string `json:"name"`
}
