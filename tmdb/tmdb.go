// Package tmdb implements communication to The Movie DB API.
// it provides function to get movie information
package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	// baseURL provides the endpoint for The Movie DB API
	baseURL = "http://api.themoviedb.org/3"
	// posterBaseURL provides the root endpoint for poster image of the movie
	posterBaseURL = "https://image.tmdb.org/t/p/w92"
)

// httpClient interface implements httpClient.Do function and intended to
// make stubbing http.Client easier during unit testing.
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

//TMDb represent an instance of tmdb API connection
type TMDb struct {
	// APIKey is required and can be acquired when we have account
	// in the movie db
	APIKey string
	// Reader is and instance of ioutil.ReadAll and defined
	// for the purpose of stubbing in unit testing and it
	// guarantee a thread safe method of stubbing.
	Reader func(io.Reader) ([]byte, error)
	// Client is an instance of httpClient interface
	Client httpClient
}

//New required tmdb API key and will return a new instance of tmdb API connection
// with all of its default setting.
func New(APIKey string) *TMDb {
	return &TMDb{
		APIKey: APIKey,
		// by default reader will use ioutil.ReadAll and client
		// will use http.Client with default timeout of 5 seconds
		Reader: ioutil.ReadAll,
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

//Result represent the information from a movie we want to
// extract from tmdb movie data.
type result struct {
	PosterPath string `json:"poster_path"`
	Overview   string `json:"overview"`
}

// searchResult represent the result from search query
// sent to tmdb API and it will extract only significant
// iformation from its original API result.
type searchResult struct {
	TotalResults int      `json:"total_results"`
	Results      []result `json:"results"`
}

//MovieInfo represent the structure of the information
// we want to get from the movie we are searching for.
type MovieInfo struct {
	Title     string
	Year      string
	Thumbnail string
	Synopsis  string
}

//GetInfo require the title and the year of searched movie
// and will send API request to TMDB API endpoint and return
// an instance of MovieInfo as result or an error is something
// wrong happened.
func (tmdb *TMDb) GetInfo(title string, year string) (*MovieInfo, error) {
	// call searchMovie to send search request to tmdb API endpoint
	result, err := tmdb.searchMovie(title, year)
	if err != nil {
		return nil, err
	}
	// if the total result 0, return error
	if result.TotalResults == 0 {
		return nil, fmt.Errorf("Couldn't find %s (%s) in TMDb", title, year)
	}
	// format the return value when a match is found
	return &MovieInfo{
		Title:     title,
		Year:      year,
		Thumbnail: fmt.Sprintf("%s%s", posterBaseURL, result.Results[0].PosterPath),
		Synopsis:  result.Results[0].Overview,
	}, nil
}

//SearchMovie will send get request to tmdb API search endpoint and will return
// an instance of searchResult.
func (tmdb *TMDb) searchMovie(title string, year string) (*searchResult, error) {
	// format the url for tmdb get request
	var URL = fmt.Sprintf("%s/search/movie?api_key=%s&query=%s&year=%s", baseURL, tmdb.APIKey, url.QueryEscape(title), year)
	// create new instance of http request
	request, _ := http.NewRequest("GET", URL, nil)
	// send get request to url defined
	res, err := tmdb.Client.Do(request)
	if err != nil {
		return nil, err
	}
	// if response status code is not 200 the return error
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP response %d", res.StatusCode)
	}
	// read body data from response
	body, err := tmdb.Reader(res.Body)
	if err != nil {
		return nil, err
	}
	// prepare target for json unmarshaling of body data
	var resp searchResult
	// extract data from body
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	// return data
	return &resp, nil
}
