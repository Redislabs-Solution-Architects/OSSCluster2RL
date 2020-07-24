package osscluster2rl

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
	Cluster string
	Server  string
}

type Cluster struct {
	Name        string
	Replication int
	KeyCount    int
	TotalMemory int
	MaxCommands int
	Nodes       []ClusterNode
	MasterNodes []string
	InitialCmd  map[string]int
	FinalCmd    map[string]int
}
