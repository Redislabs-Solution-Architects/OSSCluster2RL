package osscluster2rl

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type redishost struct {
	Host     string
	Password string
}

// Globals : global settings
type Globals struct {
	OutputFile      string
	StatsIterations int
	StatsInterval   int
}

// Config : global stuct
type Config struct {
	Clusters map[string]redishost `toml:"clusters"`
	Global   Globals              `toml:"globals"`
}

// ReadConfig : grab the whole configuration
func ReadConfig(configfile string) Config {

	var config Config

	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Unable to read config file: ", configfile)
	}

	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal("Config Error:", err)
		os.Exit(1)
	}

	return config
}
