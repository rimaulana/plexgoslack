package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
		ImageUrl: &message.Thumbnail,
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

// UpdateRepo documentation
func UpdateRepo(invoker <-chan int) {
	DataMap := make(map[int]time.Time)
	log.Println("Updater daemon status Running")
	for true {
		inv := <-invoker
		log.Println("Updater daemon invoked by watcher")
		command := fmt.Sprintf("sudo -u plex -E -H \"$LD_LIBRARY_PATH/Plex Media Scanner\" --scan --refresh --section %d", inv)
		if _, ok := DataMap[inv]; ok {
			tnow := time.Now()
			if elapsed := tnow.Sub(DataMap[inv]); elapsed >= 5 {
				DataMap[inv] = tnow
				if _, err := exec.Command("/bin/bash", "-c", command).Output(); err != nil {
					log.Println("Error :", err)
				}
			} else {
				log.Println("Updated:", elapsed, "ago")
			}
		} else {
			DataMap[inv] = time.Now()
			if _, err := exec.Command("/bin/bash", "-c", command).Output(); err != nil {
				log.Println("Error :", err)
			}
		}
	}
}

// Watcher documentation
func Watcher(root string, section int, invoker chan<- int) {
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
		isNew := false
		for _, newMovie := range diff {
			log.Println("info: detected", newMovie)
			res, err := Analyze(newMovie)
			if err != nil {
				log.Println("error:", err)
			} else {
				go PostToSlack(res)
				isNew = true
			}
		}
		if isNew {
			invoker <- section
		}
		files = files2
		time.Sleep(time.Second * 5)
	}
}

// LoadConfig documentation
func LoadConfig(RootDirectory string) (Config, error) {
	var Result Config
	log.Println("Reading config file")
	if RawConfig, err := ioutil.ReadFile(RootDirectory + "/config.toml"); err == nil {
		if _, errs := toml.Decode(string(RawConfig[:]), &Result); errs != nil {
			return Result, errs
		}
	} else {
		return Result, err
	}
	return Result, nil
}

func main() {
	// Getting configuration file
	if root, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
		if cfg, errs := LoadConfig(root); errs == nil {
			conf = cfg
		} else {
			log.Fatal("Error: ", errs)
		}
	} else {
		log.Fatal("Error: ", err)
	}

	tmdbConn = tmdb.New(conf.Tmdb.APIKey)
	invoker := make(chan int, 100)
	done := make(chan bool)

	go UpdateRepo(invoker)
	for fldr := range conf.Plex {
		go Watcher(conf.Plex[fldr].Root, conf.Plex[fldr].Section, invoker)
	}
	<-done
}
