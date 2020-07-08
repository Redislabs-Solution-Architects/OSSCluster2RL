package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	osscluster2rl "github.com/Redislabs-Solution-Architects/OSSCluster2RL/helpers"
	"github.com/go-redis/redis"
	"github.com/pborman/getopt/v2"
)

// Name is the exported name of this application.
const Name = "OSSCluster2RL"

// Version is the current version of this application.
const Version = "0.2.1"

func main() {

	var wg sync.WaitGroup
	var clusters []osscluster2rl.Cluster

	// Flags
	helpFlag := getopt.BoolLong("help", 'h', "display help")
	dbg := getopt.BoolLong("debug", 'd', "Enable debug output")
	configfile := getopt.StringLong("conf-file", 'c', "", "The path to the toml config: eg: /tmp/myconf.toml")
	getopt.Parse()

	if *helpFlag || *configfile == "" {
		getopt.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	config := osscluster2rl.ReadConfig(*configfile)

	if *dbg {
		fmt.Printf("DEBUG: Config: %+v\n", config)
	}

	rows := [][]string{
		{"name", "master_count", "replication_factor", "total_key_count", "total_memory", "maxCommands"},
	}
	csvfile, err := os.Create(config.Global.OutputFile)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	writer := csv.NewWriter(csvfile)

	for n, w := range config.Clusters {
		clusters = append(clusters)
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{w.Host},
		})
		j := rdb.ClusterNodes()
		if j.Err() != nil {
			log.Fatal("Unable to fetch clusterinformation from", w.Host)
		}
		k, parserr := osscluster2rl.ParseNodes(j)
		if parserr != nil {
			log.Fatal("Unable to get require number of nodes from: ", w.Host, ".  Run CLUSTER INFO against this node")
		}
		m := osscluster2rl.ListMasters(k)

		clusters = append(clusters,
			osscluster2rl.Cluster{
				Name:        n,
				Replication: osscluster2rl.GetReplicationFactor(k),
				KeyCount:    osscluster2rl.GetKeyspace(m, "", *dbg),
				TotalMemory: osscluster2rl.GetMemory(m, "", *dbg),
				Nodes:       k,
				MasterNodes: m,
			})
	}

	if *dbg {
		fmt.Println("DEBUG: Clusters: ", clusters)
	}

	targets := osscluster2rl.GetTargets(clusters)

	wg.Add(len(targets))
	results := make(chan osscluster2rl.CmdCount, len(targets))
	for w := 0; w < len(targets); w++ {
		go osscluster2rl.GetCommands(targets[w].Cluster, targets[w].Server, "", config.Global.StatsIterations, config.Global.StatsInterval, results, &wg, *dbg)
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
