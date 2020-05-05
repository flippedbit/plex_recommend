package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

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
	Rating          float32
	Recommendations []string
	Title           string
	imdbBody        io.Reader
}

func (movie *IMDBMovie) get_movie(id string) (*IMDBMovie, error) {
	url := imdbURL + id
	req, err := http.Get(url)
	if err != nil {
		return new(IMDBMovie), err
	}
	movie.imdbBody = req.Body
	rating, err := movie.fetchRating(movie.imdbBody)
	if err == nil {
		movie.Rating = rating
	}
	movie.fetchRecommendations()
	return movie, nil
}

/*
func (movie *IMDBMovie) fetchTitle(body *io.Reader) (string, error) {

}
*/
func (movie *IMDBMovie) fetchRating(r io.Reader) (float32, error) {
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "span" && len(t.Attr) > 0 {
				for _, att := range t.Attr {
					if att.Val == "ratingValue" {
						if z.Next() == html.TextToken {
							rating, err := strconv.ParseFloat(z.Token().Data, 64)
							if err == nil {
								rate := float32(rating)
								movie.Rating = rate
								return rate, nil
							} else {
								return 0.0, err
							}
						}
					}
				}
			}
		}
	}
	return 0.0, fmt.Errorf("Could not find rating")
}

func (movie *IMDBMovie) fetchRecommendations() {
	z := html.NewTokenizer(movie.imdbBody)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		} else if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "div" && len(t.Attr) > 0 {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "rec_overview" {
						continue
					} else if a.Key == "data-tconst" {
						if _, contain := Find(movie.Recommendations, a.Val); contain == false {
							movie.Recommendations = append(movie.Recommendations, a.Val)
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
}

func main() {
	movieID := "tt1345836"
	var movie IMDBMovie
	fmt.Println(movie.get_movie(movieID))
	/*
		url := imdbURL + movieID
		r, err := http.Get(url)
		if err != nil {
			println(err)
			os.Exit(1)
		}
		movie.imdbBody = r.Body
		z := html.NewTokenizer(movie.imdbBody)
		for {
			tt := z.Next()
			//z.Next()
			if tt == html.ErrorToken {
				break
			} else if tt == html.StartTagToken {
				t := z.Token()
				if t.Data == "div" && len(t.Attr) > 0 {
					for _, a := range t.Attr {
						if a.Key == "class" && a.Val == "rec_overview" {
							continue
						} else if a.Key == "data-tconst" {
							if _, contain := Find(movie.Recommendations, a.Val); contain == false && a.Val != movieID {
								movie.Recommendations = append(movie.Recommendations, a.Val)
							}
						} else if a.Key == "class" && a.Val == "title_wrapper" {
							if z.Next() == html.StartTagToken {
								if tn, _ := z.TagName; string(tn) == "h1" {
									if z.Next() == html.TextToken {
										fmt.Println(z.Token().Data)
									}
								}
							}
						} else {
							break
						}
					}
				} else if t.Data == "span" && len(t.Attr) > 0 {
					for _, a := range t.Attr {
						if a.Val == "ratingValue" {
							if z.Next() == html.TextToken {
								rating, err := strconv.ParseFloat(z.Token().Data, 64)
								if err == nil {
									movie.Rating = float32(rating)
								}
								//fmt.Println(t.Data)
							}
						}
					}
				}
			} else {
				continue
			}
		}
		fmt.Println(movie.Recommendations, movie.Rating)
	*/
}
