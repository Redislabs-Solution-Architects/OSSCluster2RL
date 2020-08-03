package osscluster2rl

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// GetCommands : calculate the number of commands bein process per master
func GetCommands(cluster string, server string, password string, sslConf *tls.Config, iters int, slp int, results chan<- CmdCount, wg *sync.WaitGroup, dbg bool) {
	defer wg.Done()
	prevCommands := 0
	maxCommands := 0
	client := redis.NewClient(&redis.Options{
		Addr:      server,
		Password:  password,
		TLSConfig: sslConf,
	})
	for i := 1; i <= iters; i++ {
		info := client.Info("stats")
		if dbg {
			fmt.Println("DEBUG: Fetching command count", i, "of", iters, "from", server)
			if info.Err() != nil {
				fmt.Println("Error fetching command count from ", server, "Error: ", info.Err())
			}
		}
		for _, line := range strings.Split(info.Val(), "\n") {
			r := regexp.MustCompile(`total_commands_processed:(\d+)`)
			res := r.FindStringSubmatch(line)
			if len(res) > 0 {
				j, _ := strconv.Atoi(res[1])
				if prevCommands > 0 {
					if maxCommands < j-prevCommands {
						maxCommands = j - prevCommands
						prevCommands = j
					}
				} else {
					prevCommands = j
				}
			}
		}
		time.Sleep(time.Duration(slp) * time.Second)
	}
	results <- CmdCount{Cluster: cluster, Count: maxCommands / slp}
}
