# FCM (Figg Cluster Manager)

A tool for launching Figg clusters on localhost, used for local development
and system testing.

This is inspired by [CCM (Cassandra Cluster Manager)](https://github.com/riptano/ccm)
which is used by Cassandra.

FCM is a server written in Go that manages Figg clusters, exposing a REST API
to issue commands. It runs Figg nodes in their own goroutine as part of
the FCM process.

*Note the current FCM interface is designed to support multiple node clusters,
though Figg doesn't yet support clustering so is only used for a single node
cluster at the moment.*

## Usage
FCM exposes a REST API to manage Figg clusters. Theres also a [Go SDK](./sdk).

### Cluster
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
Figg service nodes can be added with:
```
POST /clusters/{clusterID}/nodes
```
This adds a node to the cluster with the given ID, assigning a unique ID and
port to the node.

The node can then be removed again with:
```
DELETE /clusters/{clusterID}/nodes/{nodeID}
```
Which will stop the node and wait for it to exit.

### Chaos
FCM adds a [toxiproxy](https://github.com/Shopify/toxiproxy) proxy for each
node. This makes it easy to inject chaos into the cluster.

**Enable/Disable A Node**

Disable the networking for a node. This just stops proxying any traffic to the
node.
```
POST /clusters/{clusterID}/nodes/{nodeID}/disable
```

Enable the node again:
```
POST /clusters/{clusterID}/nodes/{nodeID}/enable
```

**Latency**

Add latency to the nodes network. This will return a unique handle ID.
```
POST /clusters/{clusterID}/nodes/{nodeID}/latency?latency=N
```

**Clear**

Other than enable/disable, adding chaos returns a handle ID. The chaos schenario
can be cleared with:
```
DELETE /clusters/{clusterID}/nodes/{nodeID}/chaos/{scenarioID}
```
