package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const imdbMovieURL = "https://www.imdb.com/title/"
const imdbUserURL = "https://www.imdb.com/name/"
const imdbSearchURL = "https://www.imdb.com/find?s="

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func splitIMDBName(s string) (string, error) {
	nameID := strings.Split(s, "/")
	if strings.HasPrefix(nameID[2], "nm") || strings.HasPrefix(nameID[2], "tt") {
		return nameID[2], nil
	} else {
		return "", fmt.Errorf("Could not find nameID")
	}
}

type IMDBUser struct {
	name     string
	id       string
	knownFor []string
	imdbBody string
}

type IMDBMovie struct {
	id              string
	rating          float32
	recommendations []string
	title           string
	imdbBody        string
	genre           []string
	directors       []IMDBUser
	cast            []IMDBUser
}

func newIMDBMovie() *IMDBMovie {
	return &IMDBMovie{
		title:  "",
		id:     "",
		rating: 0,
	}
}

func newIMDBUser() *IMDBUser {
	return &IMDBUser{
		name: "",
		id:   "",
	}
}

func (person *IMDBUser) GetPersonIDByName(name string) (string, error) {
	url := imdbSearchURL + "nm&q=" + strings.ReplaceAll(name, " ", "+")
	if person.imdbBody == "" {
		req, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer req.Body.Close()

		if body, err := ioutil.ReadAll(req.Body); err == nil {
			person.imdbBody = string(body)
		}
	}

	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	result := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "td" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "result_text" {
						result = true
					}
				}
			} else if t.Data == "a" && result == true {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						if href, err := splitIMDBName(attr.Val); err == nil {
							person.id = href
							return href, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("Could not find person")
}

func (person *IMDBUser) GetPersonByID(id string) error {
	person.id = id
	url := imdbUserURL + id
	if person.imdbBody == "" {
		req, err := http.Get(url)
		if err != nil {
			return err
		}
		defer req.Body.Close()

		if body, err := ioutil.ReadAll(req.Body); err == nil {
			person.imdbBody = string(body)
		} else {
			return err
		}
	}

	if person.name == "" {
		if err := person.fetchName(); err != nil {
			return err
		}
	}

	if len(person.knownFor) == 0 {
		if err := person.fetchKnownFor(); err != nil {
			return err
		}
	}

	return nil
}

func (person *IMDBUser) fetchName() error {
	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	header := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.TextToken:
			if header {
				person.name = string(z.Text())
				return nil
			} else {
				continue
			}
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "h1" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "header" {
						header = true
					}
				}
			}
		}
	}
	return nil
}

func (person IMDBUser) Name() string {
	return person.name
}

func (person *IMDBUser) fetchKnownFor() error {
	reader := strings.NewReader(person.imdbBody)
	z := html.NewTokenizer(reader)

	knownDiv := false
	depth := 0
	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "a" && len(t.Attr) > 0 {
				href := ""
				known := false
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "knownfor-ellipsis" {
						known = true
					} else if attr.Key == "href" {
						href = attr.Val
					}
				}
				if known {
					if name, err := splitIMDBName(href); err == nil {
						person.knownFor = append(person.knownFor, name)
						known = false
					}
				}
			} else if t.Data == "div" && len(t.Attr) > 0 {
				if knownDiv {
					depth++
				} else {
					for _, attr := range t.Attr {
						if attr.Key == "id" && attr.Val == "knownfor" {
							depth++
							knownDiv = true
						}
					}
				}
			}
		} else if tt == html.EndTagToken {
			if knownDiv == true {
				if tn, _ := z.TagName(); string(tn) == "div" {
					depth--
					if depth == 0 {
						break
					}
				}
			}
		}
	}
	if len(person.knownFor) == 0 {
		return fmt.Errorf("Could not find known for")
	} else {
		return nil
	}
}

func (person IMDBUser) KnownFor() []string {
	return person.knownFor
}

func (person IMDBUser) ID() string {
	return person.id
}

func (movie *IMDBMovie) GetMovieByID(id string) error {
	if movie.id == "" {
		movie.id = id
	}
	if movie.imdbBody == "" {
		url := imdbMovieURL + id
		req, err := http.Get(url)
		if err != nil {
			return err
		}
		defer req.Body.Close()

		if body, err := ioutil.ReadAll(req.Body); err == nil {
			movie.imdbBody = string(body)
		} else {
			return err
		}
	}

	if err := movie.fetchRating(); err != nil {
		return err
	}

	if err := movie.fetchTitle(); err != nil {
		return err
	}
	if err := movie.fetchRecommendations(); err != nil {
		return err
	}

	if err := movie.fetchGenre(); err != nil {
		return err
	}

	if err := movie.fetchDirectors(); err != nil {
		return err
	}

	if err := movie.fetchCast(5); err != nil {
		return err
	}

	return nil
}

func (movie *IMDBMovie) GetMovieIDByName(name string) (string, error) {
	url := imdbSearchURL + "tt&q=" + strings.ReplaceAll(name, " ", "+")
	req, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	reader := strings.NewReader(string(body))
	z := html.NewTokenizer(reader)
	result := false
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "td" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "result_text" {
						result = true
					}
				}
			} else if t.Data == "a" && result == true {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						if href, err := splitIMDBName(attr.Val); err == nil {
							movie.id = href
							return href, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("Could not find person")
}

func (movie IMDBMovie) ID() string {
	return movie.id
}

func (movie *IMDBMovie) fetchTitle() error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "h1" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "" {
						if z.Next() == html.TextToken {
							t = z.Token()
							//fmt.Println(t.Data)
							movie.title = t.Data
							return nil
						}
					}
				}
			}
		}
	}
	if movie.title == "" {
		return fmt.Errorf("Could not find title")
	} else {
		return nil
	}
}

func (movie IMDBMovie) Title() string {
	return movie.title
}

func (movie *IMDBMovie) fetchRating() error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "span" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Val == "ratingValue" {
						if z.Next() == html.TextToken {
							t = z.Token()
							rating, err := strconv.ParseFloat(t.Data, 64)
							if err == nil {
								movie.rating = float32(rating)
								return nil
							} else {
								return err
							}
						}
					}
				}
			}
		}
	}
	if movie.rating != 0 {
		return nil
	} else {
		return fmt.Errorf("Could not find rating")
	}
}

func (movie IMDBMovie) Rating() float32 {
	return movie.rating
}

func (movie *IMDBMovie) fetchRecommendations() error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "div" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "rec_overview" {
						continue
					} else if attr.Key == "data-tconst" {
						if _, contain := Find(movie.recommendations, attr.Val); contain == false {
							movie.recommendations = append(movie.recommendations, attr.Val)
						} else {
							break
						}
					}
				}
			} else {
				continue
			}
		}
	}
	if len(movie.recommendations) > 0 {
		return nil
	} else {
		return fmt.Errorf("Could not gather movie recommendations")
	}
}

func (movie IMDBMovie) Recommendations() []string {
	return movie.recommendations
}

func (movie *IMDBMovie) fetchGenre() error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			correct := false
			if t.Data == "h4" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "inline" {
						if z.Next() == html.TextToken && string(z.Text()) == "Genres:" {
							correct = true
							break
						} else {
							continue
						}
					}
				}
				if correct == true {
					for {
						tt = z.Next()
						if tt == html.ErrorToken {
							break
						} else if tt == html.StartTagToken || tt == html.EndTagToken {
							tn, _ := z.TagName()
							if string(tn) == "a" && tt == html.StartTagToken {
								if tt = z.Next(); tt == html.TextToken {
									movie.genre = append(movie.genre, strings.TrimSpace(string(z.Text())))
									continue
								}
							} else if string(tn) == "div" && tt == html.EndTagToken {
								return nil
							}
						}
					}
				}
			} else {
				continue
			}
		} else {
			continue
		}
	}
	return fmt.Errorf("Could not find genres")
}

func (movie IMDBMovie) Genre() []string {
	return movie.genre
}

func (movie *IMDBMovie) fetchDirectors() error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			correct := false
			if t.Data == "h4" && len(t.Attr) > 0 {
				for _, attr := range t.Attr {
					if attr.Key == "class" && attr.Val == "inline" {
						if z.Next() == html.TextToken && string(z.Text()) == "Director:" {
							correct = true
							break
						} else {
							continue
						}
					}
				}
				if correct == true {
					for {
						var director IMDBUser
						tt = z.Next()
						if tt == html.ErrorToken {
							break
						} else if tt == html.StartTagToken || tt == html.EndTagToken {
							t = z.Token()
							if t.Data == "a" && tt == html.StartTagToken {
								for _, attr := range t.Attr {
									if attr.Key == "href" {
										if id, err := splitIMDBName(attr.Val); err == nil {
											director.id = id
										}
									}
								}
								if tt = z.Next(); tt == html.TextToken {
									director.name = strings.TrimSpace(string(z.Text()))
									movie.directors = append(movie.directors, director)
									continue
								}
							} else if t.Data == "div" && tt == html.EndTagToken {
								return nil
							}
						}
					}
				}
			} else {
				continue
			}
		} else {
			continue
		}
	}
	return fmt.Errorf("Could not find genres")
}

func (movie IMDBMovie) Directors() []IMDBUser {
	return movie.directors
}

func (movie *IMDBMovie) fetchCast(i int) error {
	reader := strings.NewReader(movie.imdbBody)
	z := html.NewTokenizer(reader)

	depth := 0
	var castMember IMDBUser
	foundCastTable := false
	foundCastMember := false
	foundCastLink := false
	for {

		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return z.Err()
		case html.TextToken:
			if foundCastLink == true {
				if depth > 0 {
					castMember.name = strings.TrimSpace(string(z.Text()))
					movie.cast = append(movie.cast, castMember)
					foundCastMember = false
					foundCastLink = false
					castMember = IMDBUser{}
				}
			}
		case html.StartTagToken, html.EndTagToken:
			t := z.Token()
			if tt == html.StartTagToken {
				depth++
				if t.Data == "table" && len(t.Attr) > 0 {
					for _, attr := range t.Attr {
						if attr.Key == "class" && attr.Val == "cast_list" {
							foundCastTable = true
						} else {
							continue
						}
					}
				} else if t.Data == "td" && len(t.Attr) == 0 {
					if foundCastTable {
						foundCastMember = true
					}
				} else if t.Data == "a" && len(t.Attr) > 0 {
					if foundCastMember {
						for _, attr := range t.Attr {
							if attr.Key == "href" {
								if id, err := splitIMDBName(attr.Val); err == nil {
									castMember.id = id
									foundCastLink = true
								}

							}
						}
					}
				}
			} else {
				depth--
			}
		}
	}
	if len(movie.cast) > 0 {
		return nil
	} else {
		return fmt.Errorf("Could not find cast")
	}
}

func (movie IMDBMovie) Cast() []IMDBUser {
	return movie.cast
}

func main() {
	/*movieID := "tt0993846"
	movie := newIMDBMovie()
	err := movie.GetMovie(movieID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(movie.Title(), movie.ID())
	fmt.Println(movie.Recommendations())
	fmt.Println(movie.Rating())
	fmt.Println(movie.Genre())
	for _, p := range movie.Directors() {
		p.GetPerson(p.ID())
		fmt.Println(p.Name(), p.ID())
		fmt.Println(p.KnownFor())
	}
	fmt.Println("--- Cast:")
	for _, p := range movie.Cast() {
		p.GetPersonByID(p.ID())
		fmt.Println(p.Name(), p.ID())
		fmt.Println(p.KnownFor())
	}*/
	/*var p IMDBUser
	fmt.Println(p.GetPersonIDByName("Keanu Reeves"))
	fmt.Println(p.id)*/
	var m IMDBMovie
	fmt.Println(m.GetMovieIDByName("The Matrix"))
	fmt.Println(m.ID())
	m.GetMovieByID(m.ID())
	fmt.Println(m.Title())
	fmt.Println(m.Recommendations())
}
