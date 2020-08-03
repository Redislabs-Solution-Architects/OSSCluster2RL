package osscluster2rl

import (
	"strconv"
)

func GetReplicationFactor(clusterNodes []ClusterNode) int {
	var repFactor []int
	for _, v := range clusterNodes {
		if v.Role == "master" {
			repFactor = append(repFactor, len(v.Slaves))
		}
	}
	return (SliceMax(repFactor))

}

func GetTargets(c []Cluster) []CmdTarget {
	var targets []CmdTarget
	for _, w := range c {
		for _, t := range w.MasterNodes {
			targets = append(targets, CmdTarget{Cluster: w.Name, Server: t, Password: w.Password, SSL: w.SSL})
		}
	}
	return targets
}

func ListMasters(clusterNodes []ClusterNode) []string {
	var masters []string
	for _, v := range clusterNodes {
		if v.Role == "master" {
			masters = append(masters, v.IP+":"+strconv.Itoa(v.Port))
		}
	}
	return masters
}
