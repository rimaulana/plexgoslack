package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

const (
	defaultConfigPath = "config.toml"
)

// CfgLoader represent the instace of config package
type CfgLoader struct {
	Reader func(filename string) ([]byte, error)
}

// TmdbCfg represent a section on toml config file
// that hold the APIKey required to authenticate
// request to The Movie DB API endpoint.
type TmdbCfg struct {
	APIKey string `toml:"api_key"`
}

// SlackCfg represent a section on toml config file
// that contains a collection of Slack webhook that
// will be contacted on when there is new update on
// the movie collection
type SlackCfg struct {
	Webhook []string `toml:"webhooks"`
}

// PlexLibCfg represents a section on toml config file.
// it holds the information on the folder that needs to
// monitored for changes and the plex section number for
// the associated folder
type PlexLibCfg struct {
	Root    string `toml:"root"`
	Section int    `toml:"section"`
}

// Config represent the main configuration file that
// contains all sections of the config. This will be the
// one that will be the result of this package
type Config struct {
	Tmdb    TmdbCfg               `toml:"tmdb"`
	PlexURL string                `toml:"plex_url"`
	Plex    map[string]PlexLibCfg `toml:"plex"`
	Slack   SlackCfg              `toml:"slack"`
}

// New creates new instance of CfgLoader with its default
// property
func New() *CfgLoader {
	return &CfgLoader{
		Reader: ioutil.ReadFile,
	}
}

// Load will read configuration file passed as parameter
// and will parse it into Config struct
func (cfg *CfgLoader) Load(Path string) (*Config, error) {
	var configPath = defaultConfigPath
	if len(Path) >= 0 {
		configPath = Path
	}
	rawData, err := cfg.Reader(configPath)
	if err != nil {
		return nil, err
	}
	var buffer Config
	_, errs := toml.Decode(string(rawData[:]), &buffer)
	if errs != nil {
		return nil, errs
	}
	return &buffer, nil
}
