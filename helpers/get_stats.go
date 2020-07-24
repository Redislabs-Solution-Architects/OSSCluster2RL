package osscluster2rl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// GetMemory : collect memory usage information
func GetMemory(servers []string, password string, dbg bool) int {
	bytes := 0
	for _, server := range servers {
		client := redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password, // no password set
		})
		info := client.Info("memory")
		if dbg {
			fmt.Println("DEBUG: Fetching memory usage information from", server)
			if info.Err() != nil {
				fmt.Println("Error fetching memory usage data from ", server, "Error: ", info.Err())
			}
		}
		for _, line := range strings.Split(info.Val(), "\n") {
			r := regexp.MustCompile(`used_memory:(\d+)`)
			res := r.FindStringSubmatch(line)
			if len(res) > 0 {
				j, _ := strconv.Atoi(res[1])
				bytes += j
			}
		}
	}
	return (bytes)
}

// GetKeyspace : collect memory keyspace information
func GetKeyspace(servers []string, password string, dbg bool) int {
	keys := 0
	for _, server := range servers {
		client := redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password, // no password set
		})
		info := client.Info("keyspace")
		if dbg {
			fmt.Println("DEBUG: Fetching memory keyspace from", server)
			if info.Err() != nil {
				fmt.Println("Error fetching keyspace data from ", server, "Error: ", info.Err())
			}
		}
		for _, line := range strings.Split(info.Val(), "\n") {
			r := regexp.MustCompile(`db\d+:keys=(\d+),`)
			res := r.FindStringSubmatch(line)
			if len(res) > 0 {
				j, _ := strconv.Atoi(res[1])
				keys += j
			}
		}
	}
	return keys
}

// GetCmdStats : collect command stat information
func GetCmdStats(servers []string, password string, dbg bool) map[string]int {
	cmdstats := make(map[string]int)
	for _, server := range servers {
		client := redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password, // no password set
		})
		info := client.Info("commandstats")
		for _, line := range strings.Split(info.Val(), "\n") {
			r := regexp.MustCompile(`cmdstat_(\w+):calls=(\d+),`)
			res := r.FindStringSubmatch(line)
			if len(res) == 3 {
				j, _ := strconv.Atoi(res[2])
				cmdstats[res[1]] += j
			}
		}
	}
	return cmdstats
}
