package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const imdbURL = "https://www.imdb.com/title/"

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

type IMDBMovie struct {
	ID              string
	Rating          float32
	Recommendations []string
	Title           string
	imdbBody        string
	Genre           []string
}

func newIMDBMovie() *IMDBMovie {
	return &IMDBMovie{
		Title:  "",
		ID:     "",
		Rating: 0,
	}
}

func (movie *IMDBMovie) get_movie(id string) error {
	movie.ID = id
	url := imdbURL + id
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

	return nil
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
							movie.Title = t.Data
							return nil
						}
					}
				}
			}
		}
	}
	if movie.Title == "" {
		return fmt.Errorf("Could not find title")
	} else {
		return nil
	}
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
								movie.Rating = float32(rating)
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
	if movie.Rating != 0 {
		return nil
	} else {
		return fmt.Errorf("Could not find rating")
	}
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
						if _, contain := Find(movie.Recommendations, attr.Val); contain == false {
							movie.Recommendations = append(movie.Recommendations, attr.Val)
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
	if len(movie.Recommendations) > 0 {
		return nil
	} else {
		return fmt.Errorf("Could not gather movie recommendations")
	}
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
									movie.Genre = append(movie.Genre, strings.TrimSpace(string(z.Text())))
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

func main() {
	movieID := "tt1345836"
	movie := newIMDBMovie()
	err := movie.get_movie(movieID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(movie.Title, movie.ID)
	fmt.Println(movie.Recommendations)
	fmt.Println(movie.Rating)
	fmt.Println(movie.Genre)
}
