package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const imdbMovie = "https://www.imdb.com/title/"

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
	if strings.HasPrefix(nameID[2], "nm") {
		return nameID[2], nil
	} else {
		return "", fmt.Errorf("Could not find nameID")
	}
}

type IMDBUser struct {
	name     string
	id       string
	knownFor []string
}

type IMDBMovie struct {
	id              string
	rating          float32
	recommendations []string
	title           string
	imdbBody        string
	genre           []string
	directors       []IMDBUser
}

func newIMDBMovie() *IMDBMovie {
	return &IMDBMovie{
		title:  "",
		id:     "",
		rating: 0,
	}
}

func (movie *IMDBMovie) get_movie(id string) error {
	movie.id = id
	url := imdbMovie + id
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

	err = movie.fetchRating()
	if err != nil {
		return err
	}
	err = movie.fetchTitle()
	//fmt.Println(movie.Title)
	if err != nil {
		return err
	}
	err = movie.fetchRecommendations()
	if err != nil {
		return err
	}

	err = movie.fetchGenre()
	if err != nil {
		return err
	}

	err = movie.fetchDirectors()
	if err != nil {
		return err
	}

	return nil
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

func main() {
	movieID := "tt0993846"
	movie := newIMDBMovie()
	err := movie.get_movie(movieID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(movie.Title(), movie.ID())
	fmt.Println(movie.Recommendations())
	fmt.Println(movie.Rating())
	fmt.Println(movie.Genre())
	fmt.Println(movie.Directors())
}
