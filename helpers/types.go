package osscluster2rl

import "crypto/tls"

type ClusterNode struct {
	ID      string
	IP      string
	Port    int
	Cmdport int
	Role    string
	Slaves  []string
}

type CmdCount struct {
	Cluster string
	Count   int
}

type CmdTarget struct {
	Cluster  string
	Server   string
	Password string
	SSL      *tls.Config
}

type Cluster struct {
	Name        string
	Replication int
	KeyCount    int
	TotalMemory int
	MaxCommands int
	Nodes       []ClusterNode
	Password    string
	MasterNodes []string
	InitialCmd  map[string]int
	FinalCmd    map[string]int
	InitialUsec map[string]int
	FinalUsec   map[string]int
	SSL         *tls.Config
}
