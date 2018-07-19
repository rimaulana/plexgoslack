package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var (
	sampleTOML = `plex_url = "https://apps.plex.tv/"
[tmdb]
api_key = "1234567890"
[slack]
webhooks = ["slack_webhook_1","slack_webhook_2"]
[plex]
[plex.movies] # the naming after plex. is up to you
root = "/path/to/movie" #path where you keep you movie2 collection
section = 1
[plex.show] # the naming after plex. is up to you
root = "/path/to/shows" #path where you keep you movie2 collection
section = 2`
	sampleOutput = &Config{
		Tmdb: TmdbCfg{
			APIKey: "1234567890",
		},
		Slack: SlackCfg{
			Webhook: []string{"slack_webhook_1", "slack_webhook_2"},
		},
		PlexURL: `https://apps.plex.tv/`,
		Plex: map[string]PlexLibCfg{
			"movies": PlexLibCfg{
				Root:    `/path/to/movie`,
				Section: 1,
			},
			"show": PlexLibCfg{
				Root:    `/path/to/shows`,
				Section: 2,
			},
		},
	}
)

func readerStub(raw string, err error) func(strw string) ([]byte, error) {
	if err != nil {
		return func(strw string) ([]byte, error) {
			return nil, err
		}
	}
	return func(strw string) ([]byte, error) {
		return []byte(strw), nil
	}
}

var cases = []struct {
	name   string
	toml   string
	output *Config
	err    error
	errMsg string
}{
	{
		name:   "case success decoding toml",
		toml:   sampleTOML,
		output: sampleOutput,
		err:    nil,
		errMsg: "",
	},
	{
		name:   "case failed in opening the file",
		toml:   "",
		output: nil,
		err:    fmt.Errorf("couldn't find config.toml file"),
		errMsg: "couldn't find config.toml file",
	},
	{
		name:   "case failed in decoding toml",
		toml:   "\\\\",
		output: sampleOutput,
		err:    nil,
		errMsg: "bare keys cannot contain",
	},
}

func TestConfig_Load(t *testing.T) {
	for _, tt := range cases {
		cfg := New()
		cfg.Reader = readerStub(tt.toml, tt.err)
		t.Run(tt.name, func(t *testing.T) {
			result, err := cfg.Load(tt.toml)
			if err != nil {
				got := err.Error()
				want := tt.errMsg
				if !strings.Contains(got, want) {
					t.Errorf("Error in Load() = %v, want %v", got, want)
				}
			} else {
				if !reflect.DeepEqual(result, tt.output) {
					t.Errorf("Load() = %v, want %v", result, tt.output)
				}
			}
		})
	}
}
