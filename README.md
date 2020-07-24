# osscluster2rl

This program collects usage data from open source redis clusters and obtains the data necessary to size for a Redis Enterprise cluster

## Results
The data is returned in a CSV file similar to the following:

```
Cluster_Capacity

name,master_count,replication_factor,total_key_count,total_memory,maxCommands

staging,3,1,1,8235656,412087
production,33,1,1,9884703,765487716

Command_stats

cluster,command,count,usec,avg_usec_per_call

staging,set,2975648,6992772,2.35
staging,get,22317360,42402984,1.90

production,set,6284568576,14768736153,2.35
production,get,47134264320,89555102208,1.90
production,ping,211763,88941,0.42

```
| stat | description | notes |
|---|---|---|
|master_count|number of master nodes in the cluster||
|replication_factor|the number of slaves per master in the cluster||
|total_key_count|Total number of keys on the cluster|sum of key count from all master nodes|
|total_memory|Amount of memory used by Redis|sum of  used_memory from all master nodes, multiply by 2 factor if using HA in Redis Enterprise|
|maxCommands|Maximum of the Commands per Second run on the cluster over the collection period| unit = ops/second|

Command_stats are the count for each Redis command from the start of the stats gathering to the end of the run, useful for determining the operational complexity being seen across the cluster.

## Usage
0. Download the [.tar.gz binaries](https://github.com/Redislabs-Solution-Architects/OSSCluster2RL/releases) and unzip
1. copy the example_config.toml file and edit
2. Add nodes, you only need to specify a single node in the cluster the script will auto identify the rest
3. Run the binary. eg for Linux: ```./osscluster2rl_linux_amd64 -c config.toml```

```
Usage: osscluster2rl [-dh] [-c value] [parameters ...]
 -c, --conf-file=value
              The path to the toml config: eg: /tmp/myconf.toml
 -d, --debug  Enable debug output
 -h, --help   display help
```

If you have any issues, please run with the ``` -d ``` flag and submit the output as part of the bug report

## Building
0. Ensure you have go >= 1.13 and make installed on your machine
1. run ```make```
2. Run the binary. ```./osscluster2rl -c config.toml -d ```
