package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	osscluster2rl "github.com/Redislabs-Solution-Architects/OSSCluster2RL/helpers"
	"github.com/go-redis/redis"
)

type cmdCount struct {
	cluster string
	count   int
}

type clusterNode struct {
	id      string
	ip      string
	port    int
	cmdport int
	role    string
	slaves  []string
}

func listMasters(clusterNodes []clusterNode) []string {
	var masters []string
	for _, v := range clusterNodes {
		if v.role == "master" {
			masters = append(masters, v.ip+":"+strconv.Itoa(v.port))
		}
	}
	return masters
}

func parseNodes(nodes *redis.StringCmd) []clusterNode {
	var clusterNodes []clusterNode
	// the order is not set so we need to run through this loop twice first to get the masters
	for _, line := range strings.Split(nodes.Val(), "\n") {
		ln := strings.Split(line, " ")
		if len(ln) > 1 {
			r := regexp.MustCompile(`(\S+):(\d+)@(\d+)`)
			res := r.FindStringSubmatch(ln[1])
			match, _ := regexp.MatchString("master", ln[2])
			if match {
				i, _ := strconv.Atoi(res[2])
				j, _ := strconv.Atoi(res[3])
				n := clusterNode{
					id:      ln[0],
					role:    "master",
					ip:      res[1],
					port:    i,
					cmdport: j,
				}
				clusterNodes = append(clusterNodes, n)
			}
		}
	}
	// TODO: DRY this up
	for _, line := range strings.Split(nodes.Val(), "\n") {
		ln := strings.Split(line, " ")
		if len(ln) > 1 {
			r := regexp.MustCompile(`(\S+):(\d+)@(\d+)`)
			res := r.FindStringSubmatch(ln[1])
			match, _ := regexp.MatchString("slave", ln[2])

			if match {
				i, _ := strconv.Atoi(res[2])
				j, _ := strconv.Atoi(res[3])
				n := clusterNode{
					id:      ln[0],
					role:    "slave",
					ip:      res[1],
					port:    i,
					cmdport: j,
				}
				clusterNodes = append(clusterNodes, n)
				for i, v := range clusterNodes {
					if v.id == ln[3] {
						clusterNodes[i].slaves = append(clusterNodes[i].slaves, ln[0])
					}
				}
			}

		}
	}

	return clusterNodes
}

func getKeyspace(servers []string, password string) int {
	keys := 0
	for _, server := range servers {
		client := redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password, // no password set
		})
		info := client.Info("keyspace")
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

func getMemory(servers []string, password string) int {
	bytes := 0
	for _, server := range servers {
		client := redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password, // no password set
		})
		info := client.Info("memory")
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

func getCommands(cluster string, server string, password string, iters int, slp int, results chan<- cmdCount, wg *sync.WaitGroup) {
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
	results <- cmdCount{cluster: cluster, count: maxCommands / slp}
}

func getReplicationFactor(clusterNodes []clusterNode) int {
	var repFactor []int
	for _, v := range clusterNodes {
		if v.role == "master" {
			repFactor = append(repFactor, len(v.slaves))
		}
	}
	return (sliceMax(repFactor))
}

func sliceMax(s []int) int {
	m := 0
	for i, e := range s {
		if i == 0 || e > m {
			m = e
		}
	}
	return (m)
}

func main() {

	var wg sync.WaitGroup
	var configfile string

	// Read config
	flag.StringVar(&configfile, "conf", "", "path to the config file")
	flag.Parse()

	if configfile == "" {
		fmt.Println("Please sepecify a config file. Run with -h for help")
		os.Exit(1)
	}

	config := osscluster2rl.ReadConfig(configfile)

	rows := [][]string{
		{"name", "master_count", "replication_factor", "total_key_count", "total_memory", "maxCommands"},
	}
	csvfile, err := os.Create(config.Global.OutputFile)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	writer := csv.NewWriter(csvfile)

	for n, j := range config.Nodes {
		fmt.Println(n)
		fmt.Printf("%+v\n", j)
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		//Addrs: []string{config.Host},
		Addrs: []string{"localhost:30001"},
	})
	j := rdb.ClusterNodes()
	k := parseNodes(j)
	m := listMasters(k)
	wg.Add(len(m))
	results := make(chan cmdCount, len(m))
	for w := 0; w < len(m); w++ {
		go getCommands("cluster", m[w], "", config.Global.StatsIterations, config.Global.StatsInterval, results, &wg)
	}
	wg.Wait()
	close(results)
	cmds := 0
	for elem := range results {
		cmds += elem.count
	}
	r := []string{
		"localhost:30001",
		strconv.Itoa(len(m)),
		strconv.Itoa(getReplicationFactor(k)),
		strconv.Itoa(getKeyspace(m, "")),
		strconv.Itoa(getMemory(m, "")),
		strconv.Itoa(cmds)}
	rows = append(rows, r)
	for _, record := range rows {
		if err := writer.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	writer.Flush()
	os.Exit(0)
}
