package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/rimaulana/plexgoslack/config"
	"github.com/rimaulana/plexgoslack/tmdb"
)

var (
	tmdbConn   *tmdb.TMDb
	conf       *config.Config
	configPath string
)

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
	for _, hook := range conf.Slack.Webhook {
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
func Analyze(path string) (*tmdb.MovieInfo, error) {
	regex := regexp.MustCompile("((?:[^\\/]+)(?:(?:\\S+\\s+)))\\(([0-9]{4})\\)\\/?$")
	result := regex.FindStringSubmatch(path)
	if len(result) == 3 {
		res, err := tmdbConn.GetInfo(strings.TrimSpace(result[1]), strings.TrimSpace(result[2]))
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	return nil, errors.New("Path doesn't match regex")
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
				go PostToSlack(*res)
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

func init() {
	flag.StringVar(&configPath, "config", "./config.toml", "path to the specified config file, by default it is ./config.toml")
	root, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	configPath = fmt.Sprintf("%s/config.toml", root)
}

func main() {
	flag.Parse()

	// Load configuration file
	loader := config.New()
	cfg, err := loader.Load(configPath)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	conf = cfg

	tmdbConn = tmdb.New(conf.Tmdb.APIKey)
	invoker := make(chan int, 100)
	done := make(chan bool)

	go UpdateRepo(invoker)
	for fldr := range conf.Plex {
		go Watcher(conf.Plex[fldr].Root, conf.Plex[fldr].Section, invoker)
	}
	<-done
}
