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

type MediaContainer struct {
	XMLName xml.Name `xml:"MediaContainer"`
	Movies  []Movie  `xml:"Video"`
}

type Movie struct {
	XMLName  xml.Name `xml:"Video"`
	GUID     string   `xml:"guid,attr"`
	Title    string   `xml:"title,attr"`
	Rating   string   `xml:"rating,attr"`
	Added    string   `xml:"addedAt,attr"`
	Genres   []Genre  `xml:"Genre"`
	Director Director `xml:"Director"`
	Cast     []Actor  `xml:"Role"`
}

type Genre struct {
	XMLName xml.Name `xml:"Genre"`
	genre   string   `xml:"tag,attr"`
}

type Director struct {
	XMLName xml.Name `xml:"Director"`
	Name    string   `xml:"tag,attr"`
}

type Actor struct {
	XMLName xml.Name `xml:"Role"`
	Name    string   `xml:"tag,attr"`
}

// com.plexapp.agents.imdb://tt0458525?lang=en
func (m *Movie) getID() string {
	id := strings.Split(m.GUID, "//")
	id = strings.Split(id[1], "?")
	if strings.Contains(id[0], "tt") {
		return id[0]
	} else {
		return ""
	}
}

func main() {
	libraryURL := "library/sections/1/all"

	var movies MediaContainer

	url := plexURL + libraryURL + plexToken
	println(url)
	r, err := http.Get(url)
	if err != nil {
		println(err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		println(err)
		os.Exit(1)
	}

	xml.Unmarshal(body, &movies)
	for _, m := range movies.Movies {
		fmt.Println("Movie: " + m.Title + " ID: " + m.getID())
	}
}
