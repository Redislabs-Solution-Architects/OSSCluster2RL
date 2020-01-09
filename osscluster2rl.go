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

func listMasters(clusterNodes []osscluster2rl.ClusterNode) []string {
	var masters []string
	for _, v := range clusterNodes {
		if v.Role == "master" {
			masters = append(masters, v.IP+":"+strconv.Itoa(v.Port))
		}
	}
	return masters
}

func parseNodes(nodes *redis.StringCmd) []osscluster2rl.ClusterNode {
	var clusterNodes []osscluster2rl.ClusterNode
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
				n := osscluster2rl.ClusterNode{
					ID:      ln[0],
					Role:    "master",
					IP:      res[1],
					Port:    i,
					Cmdport: j,
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
				n := osscluster2rl.ClusterNode{
					ID:      ln[0],
					Role:    "slave",
					IP:      res[1],
					Port:    i,
					Cmdport: j,
				}
				clusterNodes = append(clusterNodes, n)
				for i, v := range clusterNodes {
					if v.ID == ln[3] {
						clusterNodes[i].Slaves = append(clusterNodes[i].Slaves, ln[0])
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

func getCommands(cluster string, server string, password string, iters int, slp int, results chan<- osscluster2rl.CmdCount, wg *sync.WaitGroup) {
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
	results <- osscluster2rl.CmdCount{Cluster: cluster, Count: maxCommands / slp}
}

func getReplicationFactor(clusterNodes []osscluster2rl.ClusterNode) int {
	var repFactor []int
	for _, v := range clusterNodes {
		if v.Role == "master" {
			repFactor = append(repFactor, len(v.Slaves))
		}
	}
	return (osscluster2rl.SliceMax(repFactor))

}

func getTargets(c []osscluster2rl.Cluster) []osscluster2rl.CmdTarget {
	var targets []osscluster2rl.CmdTarget
	for _, w := range c {
		for _, t := range w.MasterNodes {
			targets = append(targets, osscluster2rl.CmdTarget{Cluster: w.Name, Server: t})
		}
	}
	return targets
}

func main() {

	var wg sync.WaitGroup
	var configfile string
	var clusters []osscluster2rl.Cluster

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
		clusters = append(clusters)
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{j.Host},
		})
		j := rdb.ClusterNodes()
		k := parseNodes(j)
		m := listMasters(k)

		clusters = append(clusters,
			osscluster2rl.Cluster{
				Name:        n,
				Replication: getReplicationFactor(k),
				KeyCount:    getKeyspace(m, ""),
				TotalMemory: getMemory(m, ""),
				Nodes:       k,
				MasterNodes: m,
			})
	}

	targets := getTargets(clusters)

	wg.Add(len(targets))
	results := make(chan osscluster2rl.CmdCount, len(targets))
	for w := 0; w < len(targets); w++ {
		go getCommands(targets[w].Cluster, targets[w].Server, "", config.Global.StatsIterations, config.Global.StatsInterval, results, &wg)
	}
	wg.Wait()
	close(results)
	cmds := make(map[string]int)
	for elem := range results {
		cmds[elem.Cluster] += elem.Count
	}

	for _, c := range clusters {

		r := []string{
			c.Name,
			strconv.Itoa(len(c.MasterNodes)),
			strconv.Itoa(c.Replication),
			strconv.Itoa(c.KeyCount),
			strconv.Itoa(c.TotalMemory),
			strconv.Itoa(cmds[c.Name])}
		rows = append(rows, r)
	}
	for _, record := range rows {
		if err := writer.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	writer.Flush()
	os.Exit(0)
}
