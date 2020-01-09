package osscluster2rl

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type redishost struct {
	Host string
}

type Globals struct {
	OutputFile      string
	StatsIterations int
	StatsInterval   int
}

type Config struct {
	Nodes  map[string]redishost `toml:"nodes"`
	Global Globals              `toml:"globals"`
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

	fmt.Printf("%+v\n", config)

	return config
}
