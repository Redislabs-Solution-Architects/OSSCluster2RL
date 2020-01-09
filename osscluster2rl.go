package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	osscluster2rl "github.com/Redislabs-Solution-Architects/OSSCluster2RL/helpers"
	"github.com/go-redis/redis"
)

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
		k := osscluster2rl.ParseNodes(j)
		m := osscluster2rl.ListMasters(k)

		clusters = append(clusters,
			osscluster2rl.Cluster{
				Name:        n,
				Replication: osscluster2rl.GetReplicationFactor(k),
				KeyCount:    osscluster2rl.GetKeyspace(m, ""),
				TotalMemory: osscluster2rl.GetMemory(m, ""),
				Nodes:       k,
				MasterNodes: m,
			})
	}

	targets := osscluster2rl.GetTargets(clusters)

	wg.Add(len(targets))
	results := make(chan osscluster2rl.CmdCount, len(targets))
	for w := 0; w < len(targets); w++ {
		go osscluster2rl.GetCommands(targets[w].Cluster, targets[w].Server, "", config.Global.StatsIterations, config.Global.StatsInterval, results, &wg)
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
