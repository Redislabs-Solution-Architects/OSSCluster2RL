package osscluster2rl

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// ParseNodes : get all the nodes in a cluster
func ParseNodes(nodes *redis.StringCmd) ([]ClusterNode, error) {
	var clusterNodes []ClusterNode
	// sanity check to keep up us from chrashing, see TestParsingBroken test case
	if len(strings.Split(nodes.Val(), "\n")) < 3 {
		return clusterNodes, errors.New("Cluster requires at least 3 nodes")
	}
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

	return clusterNodes, nil
}
