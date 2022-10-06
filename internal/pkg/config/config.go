package config

import (
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

var cfg atomic.Value
var ErrConfigNotSet = errors.New("call LoadConfig before this function")

type Configuration struct {
	MongoDsn  string     `json:"mongo_dsn"`
	RedisDsn  string     `json:"redis_dsn"`
	DbName    string     `json:"db_name"`
	Relations []Relation `json:"relations"`
	Port      uint32     `json:"port"`
}

type Relation struct {
	Parent   *Entity   `json:"parent"`
	Children []Entity `json:"children"`
	On       string    `json:"on"`
	Do       string    `json:"do"`
}

type Entity struct {
	Name       string `json:"name"`
	PrimaryKey string `json:"pk"`
	ForeignKey string `json:"fk"`
}

// Get fetches the application configuration. LoadConfig must have been called
// previously for this to work.
// Use this when you need to get access to the config object at runtime
func Get() (Configuration, error) {
	c, ok := cfg.Load().(*Configuration)
	if !ok {
		return Configuration{}, ErrConfigNotSet
	}

	return *c, nil
}

// LoadConfig is used to load the configuration from either the json config file
// or the environment variables.
func LoadConfig(path string) error {
	c := Configuration{}

	if _, err := os.Stat(path); err == nil {
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		defer f.Close()

		// load config from config.json
		if err := json.NewDecoder(f).Decode(&c); err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		log.Fatal("config: cascade config.json not detected, exiting...")
	}

	cfg.Store(&c)

	return nil
}
