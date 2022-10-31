# WCM (Wombat Cluster Manager)

A tool for launching Wombat clusters on localhost, used for local development
and system testing.

This is inspired by [CCM (Cassandra Cluster Manager)](https://github.com/riptano/ccm)
which is used by Cassandra.

WCM consists of two parts. A server written in Go that manages the cluster and
proxies cluster traffic, and a client issuing commands to the manager over HTTP.
This means its easy to write new clients, such as a CLI and Go library, useful
for testing SDKs.

## Usage

### Cluster
A cluster manages a set of nodes.

Create a cluster using:
```
POST /clusters
```
This returns a unique ID for the cluster.

Then delete it again with:
```
DELETE /clusters/{clusterID}
```
This will stop all nodes in the cluster.

### Nodes
Wombat service nodes can be added with:
```
POST /clusters/{clusterID}/nodes
```
This adds a node to the cluster with the given ID, assigning a unique ID and
port to the node. It also re-compiles the wombat service before running so
theres no need to compile after making changes.

The node can then be removed again with:
```
DELETE /clusters/{clusterID}/nodes/{nodeID}
```
Which will kill the node and wait for it to exit.

### Chaos
WCM adds a [toxiproxi](https://github.com/Shopify/toxiproxy) proxy for each
node.

## TODO
- [ ] Replace gRPC with HTTP
- [ ] Add chaos commands
- [ ] Rather than compile and run in a process, run each node in its own
Goroutine where the system selects the port
