package osscluster2rl

import (
	"github.com/go-redis/redis"
	"regexp"
	"strconv"
	"strings"
)

func ParseNodes(nodes *redis.StringCmd) []ClusterNode {
	var clusterNodes []ClusterNode
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
				n := ClusterNode{
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
				n := ClusterNode{
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
