# Figg
Figg is a simple pub/sub messaging service. It only runs on a single node so
theres no fault tolerance or horizonal scaling.

*This is a work in progress project I'm just building for fun, not really intended to be used in production.*

## Components
* [`service/`](./service): Backend Figg service,
* [`sdk/`](./sdk): Go SDK client library,
* [`cli/`](./cli): Figg CLI,
* [`docs/`](./docs): Documentation on usage and architecture,
* [`tests/`](./tests): System tests,
* [`fcm/`](./fcm): Figg cluster manager.

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

Such as to check no messages are dropped when the client disconnects, can spin
up a Figg node with FCM, stream messages every 100ms from the CLI and inject
a partition into FCM for 2 seconds, repeating every 10s:
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

# Start the CLI to stream messages every 100ms. This will throw an error if it
# detects out of order messages.
./bin/cli.sh --addr 127.0.0.1:40001 stream

# Inject a partition every 10 seconds that lasts for 2 seconds.
./bin/fcm-cli.sh chaos partition --node 72c6dcb8 --duration 2 --repeat 10
```
