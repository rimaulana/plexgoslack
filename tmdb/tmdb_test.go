package tmdb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

var (
	jsonSuccess = "{\"total_results\":1,\"results\":[{\"poster_path\":\"/poster/path\",\"overview\":\"test overview\",\"title\":\"test title\"}],\"page\":1}"
	jsonEmpty   = "{\"total_results\":0,\"results\":[],\"page\":1}"
	resultInfo  = &MovieInfo{
		Title:     "test title",
		Year:      "2018",
		Thumbnail: fmt.Sprintf("%s/poster/path", posterBaseURL),
		Synopsis:  "test overview",
	}
)

func generalSet(statusCode int, body string) *http.Response {
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		StatusCode: statusCode,
	}
}

type httpClientStub struct {
	res *http.Response
	err error
}

func (cl *httpClientStub) Do(*http.Request) (*http.Response, error) {
	return cl.res, cl.err
}

var cases = []struct {
	name         string
	result       *MovieInfo
	clientStub   *httpClientStub
	errorMessage string
}{
	{
		name:   "Case successful request with movie details",
		result: resultInfo,
		clientStub: &httpClientStub{
			res: generalSet(200, jsonSuccess),
			err: nil,
		},
		errorMessage: "",
	},
	{
		name:   "Case failed request with 404 status code",
		result: nil,
		clientStub: &httpClientStub{
			res: generalSet(404, ""),
			err: nil,
		},
		errorMessage: "HTTP response 404",
	},
	{
		name:   "Case failed request contacting server",
		result: nil,
		clientStub: &httpClientStub{
			res: nil,
			err: fmt.Errorf("Timeout reached"),
		},
		errorMessage: "Timeout reached",
	},
	{
		name:   "Case failed error in json unmarshaling",
		result: nil,
		clientStub: &httpClientStub{
			res: generalSet(200, ""),
			err: nil,
		},
		errorMessage: "unexpected end of JSON input",
	},
	{
		name:   "case failed movie information not found",
		result: nil,
		clientStub: &httpClientStub{
			res: generalSet(200, jsonEmpty),
			err: nil,
		},
		errorMessage: fmt.Sprintf("Couldn't find %s (%s) in TMDb", resultInfo.Title, resultInfo.Year),
	},
}

func TestTmdb_GetInfo(t *testing.T) {
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := New("1234567890")
			db.Client = tt.clientStub
			info, err := db.GetInfo(resultInfo.Title, resultInfo.Year)
			if err != nil {
				got := err.Error()
				want := tt.errorMessage
				if !strings.Contains(got, want) {
					t.Errorf("Error in GetInfo() = %v, want %v", got, want)
				}
			} else {
				if !reflect.DeepEqual(info, tt.result) {
					t.Errorf("Error in GetInfo() = %v, want %v", info, tt.result)
				}
			}
		})
	}
}
