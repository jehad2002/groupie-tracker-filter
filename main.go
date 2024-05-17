package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Groupie struct {
	Artists  string `json:"artists"`
	Relation string `json:"relation"`
}

type Artists struct {
	ID             int                 `json:"id"`
	Image          string              `json:"image"`
	Name           string              `json:"name"`
	Members        []string            `json:"members"`
	CreationDate   int                 `json:"creationDate"`
	FirstAlbum     string              `json:"firstAlbum"`
	DatesLocations map[string][]string `json:"datesLocations"`
	Result         bool
	NameCities     map[string]bool
}

type Relation struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

var ArtistsNew []Artists

func Func() {
	var Url = "https://groupietrackers.herokuapp.com/api"
	var GroupieNew = Groupie{}
	if !Data(Url, &GroupieNew) {
		ArtistsNew[0].Result = false
		return
	}

	if !Data(GroupieNew.Artists, &ArtistsNew) {
		ArtistsNew[0].Result = false
		return
	}

	var RelationNew = Relation{}
	if !Data(GroupieNew.Relation, &RelationNew) {
		ArtistsNew[0].Result = false
		return
	}
	for index := range ArtistsNew {
		ArtistsNew[index].DatesLocations = RelationNew.Index[index].DatesLocations
	}
	ArtistsNew[0].Result = true
}

func Data(url string, val interface{}) bool {
	res, err := http.Get(url)
	if err != nil {
		return false
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	err = json.Unmarshal(body, &val)
	if err != nil {
		return false
	}
	return true
}

type Error struct {
	Str    string
	number int
}

func Err(Str string, Status int, w http.ResponseWriter, r *http.Request) {

	Info := Error{Str, Status}
	val, err := template.ParseFiles("static/templates/error.html")

	if err != nil {
		log.Println("Error when parsing a template: %s", err)
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(Status)
	err = val.ExecuteTemplate(w, "error.html", Info)
	if err != nil {
		log.Println("Error when parsing a template: %s", err)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if !ArtistsNew[0].Result {
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	if r.Method != "GET" {
		Err("405 Method Not Allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	if r.URL.Path != "/" {
		Err("404 Status Not Found", http.StatusNotFound, w, r)
		return
	}

	t, err := template.ParseFiles("static/templates/index.html")
	if err != nil {
		fmt.Println(err)
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	SortNameCities()

	err = t.ExecuteTemplate(w, "index.html", ArtistsNew)
	if err != nil {
		log.Println("Error when parsing a template:", err)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func artist(w http.ResponseWriter, r *http.Request) {
	if !ArtistsNew[0].Result {
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	if len(r.URL.Path) < 10 || r.URL.Path[:9] != "/artists/" {
		Err("400 Bad Request", http.StatusBadRequest, w, r)
		return
	}

	if r.Method != "GET" {
		Err("405 Method Not Allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	val, err := template.ParseFiles("static/templates/artist.html")
	if err != nil {
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/artists/")
	id, err1 := strconv.Atoi(name)
	if err1 != nil {
		Err("400 Bad Request", http.StatusBadRequest, w, r)
		return
	}

	if id < 1 {
		Err("400 Bad Request", http.StatusBadRequest, w, r)
		return
	}

	if id > len(ArtistsNew) {
		Err("404 Not Found", http.StatusNotFound, w, r)
		return
	}

	err = val.ExecuteTemplate(w, "artist.html", ArtistsNew[id-1])
	if err != nil {
		log.Println("Error when parsing a template: %s", err)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func Filter(w http.ResponseWriter, r *http.Request) {

	ArtistsFilter := []Artists{}

	if !ArtistsNew[0].Result {
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	if r.URL.Path != "/filters/" {
		Err("400 Bad Request", http.StatusBadRequest, w, r)
		return
	}

	if r.Method != "GET" {
		Err("405 Method Not Allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	CreationDate := r.FormValue("CreationDate")
	if CreationDate == "on" {
		CreationDateFrom := r.FormValue("CreationDateFrom")
		CreationDateTo := r.FormValue("CreationDateTo")

		if CreationDateFrom == "" {
			CreationDateFrom = "0"
		}

		if CreationDateTo == "" {
			CreationDateTo = "2111"
		}

		CDF, err1 := strconv.Atoi(CreationDateFrom)
		CDT, err2 := strconv.Atoi(CreationDateTo)
		if err1 != nil || err2 != nil {
			Err("400 Bad Request", http.StatusBadRequest, w, r)
			return
		}

		if !CheckValue(CDF, CDT) {
			Err("400 Bad Request", http.StatusBadRequest, w, r)
			return
		}

		ArtistsFilter = CheckOnCreationDate(ArtistsFilter, CDF, CDT)
		if len(ArtistsFilter) == 0 {
			Err("Not Found", http.StatusOK, w, r)
			return
		}
	}

	FirstAlbumDate := r.FormValue("FirstAlbumDate")
	var err1 bool
	if FirstAlbumDate == "on" {
		FirstAlbumDateFrom := r.FormValue("FirstFrom")
		FirstAlbumDateTo := r.FormValue("FirstTo")

		if FirstAlbumDateFrom == "" {
			FirstAlbumDateFrom = "20-01-1000"
		}

		if FirstAlbumDateTo == "" {
			FirstAlbumDateTo = "20-01-2111"
		}

		if !CheckValueDate(FirstAlbumDateFrom, FirstAlbumDateTo) {
			Err("400 Bad Request", http.StatusBadRequest, w, r)
			return
		}

		ArtistsFilter, err1 = CheckFirstAlbumDate(ArtistsFilter, FirstAlbumDateFrom, FirstAlbumDateTo)
		if !err1 {
			Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
			return
		}

		if len(ArtistsFilter) == 0 {
			Err("Not Found", http.StatusOK, w, r)
			return
		}
	}

	NumberOfMembers := r.FormValue("NOM")
	if NumberOfMembers == "on" {

		NumberOfMembersFrom := r.FormValue("NOMfrom")
		NumberOfMembersTo := r.FormValue("NOMto")
		if NumberOfMembersFrom == "" {
			NumberOfMembersFrom = "1"
		}

		if NumberOfMembersTo == "" {
			NumberOfMembersTo = "111"
		}

		NOMF, err1 := strconv.Atoi(NumberOfMembersFrom)
		NOMT, err2 := strconv.Atoi(NumberOfMembersTo)
		if err1 != nil || err2 != nil {
			Err("400 Bad Request", http.StatusBadRequest, w, r)
			return
		}

		if !CheckValue(NOMF, NOMT) {
			Err("400 Bad Request", http.StatusBadRequest, w, r)
			return
		}

		ArtistsFilter = CheckOnNumberOfMembers(ArtistsFilter, NOMF, NOMT)
		if len(ArtistsFilter) == 0 {
			Err("Not Found", http.StatusOK, w, r)
			return
		}
	}

	LocationOfConcerts := r.FormValue("LocationOfConcerts")
	if LocationOfConcerts == "on" {
		LocationOfConcertsValue := r.FormValue("LOC")

		if LocationOfConcertsValue != "" {
			ArtistsFilter = CheckOnLocationOfConcerts(ArtistsFilter, LocationOfConcertsValue)
			if len(ArtistsFilter) == 0 {
				Err("Not Found", http.StatusOK, w, r)
				return
			}
		}
	}

	val, err := template.ParseFiles("static/templates/filter.html")
	if err != nil {
		Err("500 Internal Server Error", http.StatusInternalServerError, w, r)
		return
	}

	if len(ArtistsFilter) == 0 {
		err = val.ExecuteTemplate(w, "filter.html", ArtistsNew)
		if err != nil {
			log.Println("Error when parsing a template: %s", err)
			fmt.Fprintf(w, err.Error())
			return
		}
		return
	}

	err = val.ExecuteTemplate(w, "filter.html", ArtistsFilter)
	if err != nil {
		log.Println("Error when parsing a template: %s", err)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func HandleFuncOwn() {
	http.HandleFunc("/", index)
	http.HandleFunc("/artists/", artist)
	http.HandleFunc("/filters/", Filter)
	log.Println(http.ListenAndServe(":8080", nil))
}

func SortNameCities() {

	ArtistsNew[0].NameCities = make(map[string]bool)
	for _, value := range ArtistsNew {
		for key := range value.DatesLocations {
			ArtistsNew[0].NameCities[key] = true
		}
	}
}

func CheckValue(From, To int) bool {

	if From > To {
		return false
	}

	if From < 0 || To < 0 {
		return false
	}
	return true
}

func CheckOnCreationDate(ArtistFilter []Artists, From, To int) []Artists {

	if len(ArtistFilter) > 0 {
		FilterInFilter := []Artists{}
		for _, value := range ArtistFilter {

			if From <= value.CreationDate && value.CreationDate <= To {
				FilterInFilter = append(FilterInFilter, value)
			}
		}
		return FilterInFilter
	}

	for _, value := range ArtistsNew {

		if From <= value.CreationDate && value.CreationDate <= To {
			ArtistFilter = append(ArtistFilter, value)
		}
	}

	return ArtistFilter
}

func CheckOnNumberOfMembers(ArtistFilter []Artists, From, To int) []Artists {

	if len(ArtistFilter) > 0 {
		FilterInFilter := []Artists{}
		for _, value := range ArtistFilter {

			if From <= len(value.Members) && len(value.Members) <= To {
				FilterInFilter = append(FilterInFilter, value)
			}
		}
		return FilterInFilter
	}

	for _, value := range ArtistsNew {

		if From <= len(value.Members) && len(value.Members) <= To {
			ArtistFilter = append(ArtistFilter, value)
		}
	}

	return ArtistFilter
}

func CheckValueDate(From, To string) bool {

	from := strings.Split(From, "-")
	to := strings.Split(To, "-")

	if len(from) != 3 || len(to) != 3 {
		return false
	}

	if !checkDate(from) {
		return false
	}

	if !checkDate(to) {
		return false
	}

	return true
}

func checkDate(array []string) bool {

	Day, err := strconv.Atoi(array[0])
	if err != nil {
		return false
	}

	if Day > 31 || Day < 1 {
		return false
	}

	Month, err := strconv.Atoi(array[1])
	if err != nil {
		return false
	}

	if Month > 12 || Month < 1 {
		return false
	}

	if Month == 4 || Month == 6 || Month == 9 || Month == 11 {
		if Day > 30 {
			return false
		}
	}

	Year, err := strconv.Atoi(array[2])
	if err != nil {
		return false
	}

	if Month == 2 && Year%4 == 0 {
		if Day > 29 {
			return false
		}
	}

	if Month == 2 && Year%4 != 0 {
		if Day > 28 {
			return false
		}
	}

	if Year > 3000 || Year < 1 {
		return false
	}
	return true
}

func CheckFirstAlbumDate(ArtistFilter []Artists, From, To string) ([]Artists, bool) {

	from := strings.Split(From, "-")
	to := strings.Split(To, "-")

	fr, err := separationArray(from)
	if !err {
		return []Artists{}, false
	}

	t, err := separationArray(to)
	if !err {
		return []Artists{}, false
	}

	if len(ArtistFilter) > 0 {
		FilterInFilter := []Artists{}
		for _, value := range ArtistFilter {

			FAD := strings.Split(value.FirstAlbum, "-")
			fad, err := separationArray(FAD)
			if !err {
				return []Artists{}, false
			}

			if comparison(fad, fr, t) {
				FilterInFilter = append(FilterInFilter, value)
			}
		}
		return FilterInFilter, true
	}

	for _, value := range ArtistsNew {

		FAD := strings.Split(value.FirstAlbum, "-")
		fad, err := separationArray(FAD)
		if !err {
			return []Artists{}, false
		}

		if comparison(fad, fr, t) {
			ArtistFilter = append(ArtistFilter, value)
		}
	}

	return ArtistFilter, true
}

func comparison(fad, fr, t []int) bool {
	Fad := time.Month(fad[1])
	from := time.Month(fr[1])
	to := time.Month(t[1])

	FAD := time.Date(fad[2], Fad, fad[0], 0, 0, 0, 0, time.UTC)
	FROM := time.Date(fr[2], from, fr[0], 0, 0, 0, 0, time.UTC)
	TO := time.Date(t[2], to, t[0], 0, 0, 0, 0, time.UTC)

	From := FROM.Before(FAD)
	To := TO.After(FAD)
	EqualFrom := FROM.Equal(FAD)
	EqualTo := TO.Equal(FAD)

	if EqualFrom == true || EqualTo == true {
		return true
	}

	if From == true && To == true {
		return true
	}

	return false
}

func checkOn(ArtistFilter []Artists, value Artists) bool {
	for _, val := range ArtistFilter {
		if val.Name == value.Name {
			return true
		}
	}
	return false
}

func separationArray(array []string) ([]int, bool) {
	arrayInt := []int{}
	for _, value := range array {
		Int, err := strconv.Atoi(value)
		if err != nil {
			return []int{}, false
		}
		arrayInt = append(arrayInt, Int)
	}
	return arrayInt, true
}

func CheckOnLocationOfConcerts(ArtistFilter []Artists, Location string) []Artists {
	if len(ArtistFilter) > 0 {
		FilterInFilter := []Artists{}
		for _, value := range ArtistFilter {

			for key := range value.DatesLocations {
				Location = strings.Replace(Location, ", ", "-", 1)
				Location = strings.Replace(Location, " ", "_", -1)
				if strings.Contains(key, strings.ToLower(Location)) {
					FilterInFilter = append(FilterInFilter, value)
					break
				}
			}
		}
		return FilterInFilter
	}

	for _, value := range ArtistsNew {

		for key := range value.DatesLocations {
			Location = strings.Replace(Location, ", ", "-", 1)
			Location = strings.Replace(Location, " ", "_", -1)
			if strings.Contains(key, strings.ToLower(Location)) {
				ArtistFilter = append(ArtistFilter, value)
				break
			}
		}
	}
	return ArtistFilter
}
func main() {
	Func()
	HandleFuncOwn()
}
