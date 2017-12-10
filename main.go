package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/rimaulana/plexgoslack/tmdb"
)

var tmdbConn *tmdb.TMDb
var conf Config

// MDb explanation
type MDb struct {
	APIKey string `toml:"api_key"`
}

// SlackCfg explanation
type SlackCfg struct {
	WebHook []string `toml:"webhooks"`
}

// PlexLibrary explanation
type PlexLibrary struct {
	Root    string `toml:"root"`
	Section int    `toml:"section"`
}

// Config explanation
type Config struct {
	Tmdb    MDb                    `toml:"tmdb"`
	PlexURL string                 `toml:"plex_url"`
	Plex    map[string]PlexLibrary `toml:"plex"`
	Slack   SlackCfg               `toml:"slack"`
}

// PostToSlack documentation
func PostToSlack(message tmdb.MovieInfo) {
	text := fmt.Sprintf("New movie is now available on <%sweb/index.html|Plex>", conf.PlexURL)
	test := fmt.Sprintf("%s (%s)", message.Title, message.Year)
	head := "Synopsis"
	atth1 := slack.Attachment{
		Title:    &test,
		ImageUrl: &message.Tumbnail,
	}
	atth2 := slack.Attachment{
		Title: &head,
		Text:  &message.Synopsis,
	}
	payload := slack.Payload{
		Text:        text,
		Attachments: []slack.Attachment{atth1, atth2},
	}
	for _, hook := range conf.Slack.WebHook {
		err := slack.Send(hook, "", payload)
		log.Println("Send", message.Title, "info to Slack")
		if len(err) > 0 {
			log.Printf("error: %s\n", err)
		}
	}
}

// Diff documentation
func Diff(a, b []os.FileInfo) []string {
	mb := map[string]bool{}
	for _, x := range a {
		mb[x.Name()] = true
	}
	ab := []string{}
	for _, x := range b {
		if _, ok := mb[x.Name()]; !ok {
			ab = append(ab, x.Name())
		}
	}
	return ab
}

// Analyze documentation
func Analyze(path string) (tmdb.MovieInfo, error) {
	regex := regexp.MustCompile("((?:[^\\/]+)(?:(?:\\S+\\s+)))\\(([0-9]{4})\\)\\/?$")
	result := regex.FindStringSubmatch(path)
	if len(result) == 3 {
		res, err := tmdbConn.GetInfo(strings.TrimSpace(result[1]), strings.TrimSpace(result[2]))
		if err != nil {
			return tmdb.MovieInfo{}, err
		}
		return res, nil
	}
	return tmdb.MovieInfo{}, errors.New("Path doesn't match regex")
}

// Watcher documentation
func Watcher(root string) {
	log.Println("info: monitoring folder", root)
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	for {
		files2, err := ioutil.ReadDir(root)
		if err != nil {
			log.Println("error:", err)
		}
		diff := Diff(files, files2)
		for _, newMovie := range diff {
			log.Println("info: detected", newMovie)
			res, err := Analyze(newMovie)
			if err != nil {
				log.Println("error:", err)
			} else {
				go PostToSlack(res)
			}
		}
		files = files2
		time.Sleep(time.Second * 5)
	}
}

func main() {
	RootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("Error: ", err)
	}
	RawConfig, err := ioutil.ReadFile(RootDir + "\\config.toml")
	if err != nil {
		log.Fatal("Error: ", err)
	}
	if _, err := toml.Decode(string(RawConfig[:]), &conf); err != nil {
		log.Fatal("Error: ", err)
	}
	tmdbConn = tmdb.New(conf.Tmdb.APIKey)

	done := make(chan bool)
	for fldr := range conf.Plex {
		go Watcher(conf.Plex[fldr].Root)
	}
	<-done
}
