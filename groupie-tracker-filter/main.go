package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var artistDetails Details

func main() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/moreinfo", secPage)
	fmt.Println("Server starting on localhost:8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./template.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var artists []AR
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artists[20].Image = "https://brasilia.deboa.com/wp-content/uploads/2023/12/Mamonas-Assassinas-O-Filme.jpg"

	respo, err := http.Get("https://groupietrackers.herokuapp.com/api/locations")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer respo.Body.Close()

	var locs AllLocation
	bodyl, err := ioutil.ReadAll(respo.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bodyl, &locs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse and apply filters
	creationDateStart := r.URL.Query().Get("creationDateStart")
	creationDateEnd := r.URL.Query().Get("creationDateEnd")
	firstAlbumStart := r.URL.Query().Get("firstAlbumStart")
	firstAlbumEnd := r.URL.Query().Get("firstAlbumEnd")
	members := r.URL.Query().Get("members")
	location := r.URL.Query().Get("location")

	filteredArtists := filterArtists(artists, locs, creationDateStart, creationDateEnd, firstAlbumStart, firstAlbumEnd, members, location)

	err = tmpl.Execute(w, filteredArtists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func filterArtists(artists []AR, locs AllLocation, creationDateStart, creationDateEnd, firstAlbumStart, firstAlbumEnd, members, location string) []AR {
	var filtered []AR
	for _, artist := range artists {
		// Apply filters
		match := true

		if creationDateStart != "" && creationDateEnd != "" {
			cdStart, errStart := strconv.Atoi(creationDateStart)
			cdEnd, errEnd := strconv.Atoi(creationDateEnd)
			if errStart != nil || errEnd != nil || artist.Creation < cdStart || artist.Creation > cdEnd {
				match = false
			}
		}

		if firstAlbumStart != "" && firstAlbumEnd != "" {
			faStart, errStart := strconv.Atoi(firstAlbumStart)
			faEnd, errEnd := strconv.Atoi(firstAlbumEnd)
			firstAlbumYear, err := strconv.Atoi(artist.FirstAlbum[6:])
			if errStart != nil || errEnd != nil || err != nil || firstAlbumYear < faStart || firstAlbumYear > faEnd {
				match = false
			}
		}

		if members != "" {
			m, err := strconv.Atoi(members)
			if err != nil || len(artist.Members) != m {
				match = false
			}
		}

		if location != "" {
			found := false
			queryLocation := strings.ToLower(location)
			for _, loc := range locs.Index[artist.ID-1].Locationss {
				if strings.Contains(strings.ToLower(loc), queryLocation) {
					found = true
					break
				}
			}
			if !found {
				match = false
			}
		}

		if match {
			filtered = append(filtered, artist)
		}
	}

	return filtered
}

var found bool

func secPage(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("artistNumber")
	fetchArtistDetails(id)
	tmpl, err := template.ParseFiles("artistPage.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, artistDetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func fetchArtistDetails(id string) {
	fetchJSON("https://groupietrackers.herokuapp.com/api/artists/"+id, &artistDetails.Artist)
	fetchJSON("https://groupietrackers.herokuapp.com/api/locations/"+id, &artistDetails.Locations)
	fetchJSON("https://groupietrackers.herokuapp.com/api/dates/"+id, &artistDetails.ConcertDates)
	fetchJSON("https://groupietrackers.herokuapp.com/api/relation/"+id, &artistDetails.DatesAndLocations)
}

func fetchJSON(url string, target interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching data from %s: %v\n", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
	}
}
