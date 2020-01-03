package osscluster2rl

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Host            string
	OutputFile      string
	StatsIterations int
	StatsInterval   int
}

func ReadConfig(configfile string) Config {

	var config Config

	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Unable to read config file: ", configfile)
	}

	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	return config
}
