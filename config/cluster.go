package config

import (
	"fmt"
	"os"
)

type Cluster struct {
	// NodeName is the unique identifier for the node in cluster.
	NodeName string `yaml:"node_name"`
	// InternalAddr is the address that the cluster internal communication server will listen on.
	InternalAddr string `yaml:"internal_addr"`
	// GossipAddr is the address that the gossip will listen on, It is used for both UDP and TCP gossip.
	GossipAddr string `yaml:"gossip_addr"`

	Join []string `yaml:"join"`
}

func init() {
	hostName, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("get Hostname failed: %s", err))
	}
	DefaultCluster.NodeName = hostName
}

var (
	DefaultCluster = Cluster{
		InternalAddr: ":4456",
		GossipAddr:   ":4456",
	}
)
