package tmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const baseURL string = "http://api.themoviedb.org/3"

//TMDb 1234
type TMDb struct {
	APIKey string
}

//New bla bla bla
func New(APIKey string) *TMDb {
	return &TMDb{APIKey: APIKey}
}

//Result bla bla bla
type Result struct {
	VoteCount        int     `json:"vote_count"`
	ID               int     `json:"id"`
	Video            bool    `json:"video"`
	VoteAverage      float32 `json:"vote_average"`
	Title            string  `json:"title"`
	Popularity       float32 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	GenreIds         []int   `json:"genre_ids"`
	BackdropPath     string  `json:"backdrop_path"`
	Adult            bool    `json:"adult"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
}

//SearchResult explanation
type SearchResult struct {
	Page         int      `json:"page"`
	TotalResults int      `json:"total_results"`
	TotalPages   int      `json:"total_pages"`
	Results      []Result `json:"results"`
}

//MovieInfo explanation
type MovieInfo struct {
	Title     string
	Year      string
	Thumbnail string
	Synopsis  string
}

//GetInfo explanation
func (tmdb *TMDb) GetInfo(title string, year string) (MovieInfo, error) {
	var resp MovieInfo
	result, err := tmdb.SearchMovie(title, year)
	if err != nil {
		return resp, err
	}
	if result.TotalResults == 0 {
		return resp, errors.New("Couldn't find " + title + " (" + year + ") at TMDb")
	}
	return MovieInfo{Title: title, Year: year, Thumbnail: "https://image.tmdb.org/t/p/w92" + result.Results[0].PosterPath, Synopsis: result.Results[0].Overview}, nil
}

//SearchMovie explanation
func (tmdb *TMDb) SearchMovie(title string, year string) (SearchResult, error) {
	var resp SearchResult
	res, err := http.Get(baseURL + "/search/movie?api_key=" + tmdb.APIKey + "&query=" + url.QueryEscape(title) + "&year=" + year)
	if err != nil {
		return resp, err
	}
	if res.StatusCode != 200 {
		return resp, fmt.Errorf("HTTP Resposnse %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return SearchResult{}, err
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return SearchResult{}, err
	}
	return resp, nil
}
