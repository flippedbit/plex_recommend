package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const plexURL = "http://192.168.1.88:32400/"
const plexToken = "?X-Plex-Token=FFH5BBKsCW3iSgnDxynW"

type PlexMovieLibrary struct {
	XMLName xml.Name    `xml:"MediaContainer"`
	Movies  []PlexMovie `xml:"Video"`
}

type PlexMovie struct {
	XMLName  xml.Name     `xml:"Video"`
	GUID     string       `xml:"guid,attr"`
	Title    string       `xml:"title,attr"`
	Rating   string       `xml:"rating,attr"`
	Added    string       `xml:"addedAt,attr"`
	Genres   []PlexGenre  `xml:"Genre"`
	Director PlexDirector `xml:"Director"`
	Cast     []PlexActor  `xml:"Role"`
}

type PlexGenre struct {
	XMLName xml.Name `xml:"Genre"`
	genre   string   `xml:"tag,attr"`
}

type PlexDirector struct {
	XMLName xml.Name `xml:"Director"`
	Name    string   `xml:"tag,attr"`
}

type PlexActor struct {
	XMLName xml.Name `xml:"Role"`
	Name    string   `xml:"tag,attr"`
}

// com.plexapp.agents.imdb://tt0458525?lang=en
func (m *PlexMovie) getID() (string, error) {
	id := strings.Split(m.GUID, "//")
	id = strings.Split(id[1], "?")
	if strings.Contains(id[0], "tt") {
		return id[0], nil
	} else {
		return "", fmt.Errorf("Could not get ID")
	}
}

func main() {
	libraryURL := "library/sections/1/all"

	var movies PlexMovieLibrary

	url := plexURL + libraryURL + plexToken
	r, err := http.Get(url)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err)
		os.Exit(1)
	}

	xml.Unmarshal(body, &movies)
	for _, m := range movies.Movies {
		if id, err := m.getID(); err == nil {
			fmt.Println("Movie: " + m.Title + " ID: " + id)
		}
	}
}
