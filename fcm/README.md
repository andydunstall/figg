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

## Components
* [`server/`](./server): FCM backend server,
* [`sdk/`](./sdk): FCM Go SDK,
* [`cli/`](./cli): FCM CLI.

## Usage
This shows the CLI commands to interract with FCM. Theres also a [Go SDK](./sdk)
used for system tests, and FCM exposed a REST API to create new clients.

Note can call with `bin/fcm-cli.sh` to avoid having to recompile.

These commands require FCM to be running (`bin/fcm.sh`).

### Cluster
Create a cluster, which starts a single Figg node:
```bash
$ fcm cluster create

    Cluster
    -------
    ID: 2615f9d2

    Nodes
    -------
    ID:  72c6dcb8 | Addr: 127.0.0.1:40000 | Proxy Addr: 127.0.0.1:40001

```

The proxy address is a FCM proxy targetted at the node address, used to inject
faults into the connection. Such as may want to test subscribers resume
correctly when they disconnect, so can target a publisher at the node address
and a subscriber at the proxy address, then inject faults into the subscribers
connection.

### Chaos
FCM adds a proxy for each node. This is used in inject faults into the nodes
connections, when connected to the proxy address.

Each chaos command has options arguments:
* `duration`: How long the fault should last in seconds (if not defined lasts
forever),
* `repeat`: How often the fault should be injected in seconds (if not defined
never repeats)

#### Add Partition
Disables the proxy such that existing connections close and new connections
time out.

```
$ fcm chaos partition --node {node ID} [--duration {duration}] [--repeat {repeat}]
```

Such as can to a partition that lasts 2 seconds and repeats every 10 seconds.
```
$ fcm chaos partition --node 72c6dcb8 --duration 2 --repeat 10
```
