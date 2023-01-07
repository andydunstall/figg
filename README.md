# Figg
Figg is a simple pub/sub messaging service.

This is just a project I'm building for fun and practice Go systems performance,
so is missing lots of features needed to run in production, and only supports
a single node so no fault tolerance or partitioning yet.

## Features
* Message continuity: If subscribers connections drop they automatically
reconnect and resume any missed messages,
* Message history: Messages are persisted to a commit log on disk so subscribers
can start subscribing from an old message.

## Components
* [`server/`](./server): Backend Figg server,
* [`sdk/`](./sdk): Go SDK client library,
* [`cli/`](./cli): Figg CLI,
* [`bench/`](./bench): Figg benchmarking client,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`tests/`](./tests): System tests,
* [`fcm/`](./fcm): Figg cluster manager, used for system tests and manual chaos
testing.

## Usage
### Server
The Figg [`server`](./server) can be started with `./bin/figg-server` for local
testing (which uses `go run`) or compile the package in [`./server`](./server).

For now all configuration is passed via the command line, whose options can be
seen with `server -h`.

### Client
See [`sdk/`](./sdk) for full usage.

```go
// Connect with default options.
client, err := figg.Connect("10.26.104.52:8119")
if err != nil {
	// handle err
}

// Subscribe. Blocks until the server confirms the subscription is setup.
err := client.Subscribe("foo", func(m figg.Message)) {
	fmt.Println("message: ", string(m.Data), "offset: ", m.Offset)
})
if err != nil {
	// handle err
}

// Publish message bar to topic foo.
client.Publish("foo", []byte("bar"), func() {
	fmt.Println("message acked")
})
```

## Benchmarking
Benchmarks against the Figg service are run with `bin/figg-bench` (see
[`bench/`](./bench) for details).

Some performance critical components also have Go benchmark tests ran with
`go test` or `./build/bench.sh`.

## Testing
The server and SDK have high unit test coverage included alongside the packages
using `go test`.

Some end-to-end system tests are needed to:
* Check components are properly integrated,
* Inject chaos into a cluster.

These system tests are in [`tests/`](./tests). [`FCM`](./fcm) is used to create
Figg clusters locally, which proxies network traffic to inject chaos.

### Manual Testing
Theres tools in [`cli/`](./cli) and [`fcm/`](./fcm) to easily spin up notes
and inject chaos.

Such can use FCM to add a node and drop the network for 2 seconds every
10 seconds with:
```bash
# Create a Figg node.
$ ./bin/fcm-cli cluster create

    Cluster
    -------
    ID: 2615f9d2

    Nodes
    -------
    ID:  72c6dcb8 | Addr: 127.0.0.1:40000 | Proxy Addr: 127.0.0.1:40001

# Inject a partition every 10 seconds that lasts for 2 seconds.
$ ./bin/fcm-cli chaos partition --node 72c6dcb8 --duration 2 --repeat 10
```

Then use the CLI to stream messages, pointing the subscriber at the proxied
server which will drop networking, and the publisher at the non-proxied server
so messages continue to be published even while the subscriber is disconnected.
`figg-cli stream` will check the next message follows from the previous message
without dropping (by comparing offsets).
```bash
# Start the CLI to stream messages every 10ms. This will throw an error if it
# detects out of order messages.
$ ./bin/cli stream --sub-addr 127.0.0.1:40001 --pub-addr 127.0.0.1:40000
pub state CONNECTED
sub state CONNECTED
received 50
received 100
received 150
sub state DISCONNECTED
sub state CONNECTED
received 200
received 250
...
```
