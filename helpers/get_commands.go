package osscluster2rl

import (
	"github.com/go-redis/redis"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetCommands(cluster string, server string, password string, iters int, slp int, results chan<- CmdCount, wg *sync.WaitGroup) {
	defer wg.Done()
	prevCommands := 0
	maxCommands := 0
	client := redis.NewClient(&redis.Options{
		Addr:     server,
		Password: password, // no password set
	})
	for i := 1; i <= iters; i++ {
		info := client.Info("stats")
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
