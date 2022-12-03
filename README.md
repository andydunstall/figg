# Figg
Figg is a simple pub/sub messaging service.

**WIP**: This is only a projects I'm building for run, and to practice Go and
systems performance.

**Features**
* Resume: Subscribers can resume from an old offset (either to fetch history
or ensure message continuity across connection drops),
* Message retention: Messages are persisted to a commit log so subscribers can
resume from an old offset.

**Limitations**
* Only runs on a single node, so theres no faults tolerance or horizonal
scaling,
* Many areas are still WIP. The basic service and resume works other areas
(especially in the SDK) still need work.

## Components
* [`service/`](./service): Backend Figg service,
* [`sdk/`](./sdk): Go SDK client library,
* [`cli/`](./cli): Figg CLI,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`tests/`](./tests): System tests,
* [`fcm/`](./fcm): Figg cluster manager, used for system tests and chaos testing.

## Usage
### Service
The [`Figg service`](./service) can be started with `./bin/figg.sh`, or compile
the package in [`./service`](./service) with `go build ./...`. For now all
configuration is passed via the command line.

### SDK
See [`sdk/`](./sdk) for full usage.

#### Connect
Connects to a figg node.

```go
config := &Config{
	Addr: "mynode",
}
client := figg.NewFigg(config)
```

#### Publish
Publishes a message to topic `foo`.

```go
client.Publish("foo", []byte("bar"))
```

#### Subscribe
Subscribes to topic `foo`.

```go
sub := client.Subscribe("foo", func(topic string, m []byte) {
  fmt.Println("received message", string(m))
})
defer client.Unsubscribe("foo", sub)
```

## Testing
The service and SDK aims for high unit test coverage where possible which are
included in the [`service/`](./service) and [`sdk`](./sdk) packages alongside
the code itself.

Though some end-to-end system tests are needed to:
* Check components are properly integrated,
* Inject chaos into a cluster to check for issues overlooked in the design.
These tests are in [`tests/`](./tests). [`FCM`](./fcm) is used to create Figg
clusters locally and inject chaos, which is used both for testing the service
and the SDK.

### Manual Testing
Although most behaviours should have automated tests its often useful to run
tests manually. Theres tools in [`cli/`](./cli) and [`fcm/`](./fcm) to make
this easy.

Such as to check no messages are dropped when the subscriber disconnects, can
spin up a Figg node with FCM, inject a partition into the nodes proxy for
2 seconds every 10 seconds, and stream messages from a subscriber connected
to the proxy.
```bash

# Start FCM
./bin/fcm.sh

# Create a Figg node.
$ ./bin/fcm-cli.sh cluster create

    Cluster
    -------
    ID: 2615f9d2

    Nodes
    -------
    ID:  72c6dcb8 | Addr: 127.0.0.1:40000 | Proxy Addr: 127.0.0.1:40001

# Inject a partition every 10 seconds that lasts for 2 seconds.
$ ./bin/fcm-cli.sh chaos partition --node 72c6dcb8 --duration 2 --repeat 10

# Start the CLI to stream messages every 100ms. This will throw an error if it
# detects out of order messages.
$ ./bin/cli.sh stream --sub-addr 127.0.0.1:40001 --pub-addr 127.0.0.1:40000
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
