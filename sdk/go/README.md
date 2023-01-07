# Figg Go SDK

## Usage
### Connect
Note though `Connect` waits for the initial connection to succeed, if the
connection drops the SDK will automatically reconnect.

```go
import (
	"github.com/dunstall/figg/sdk"
)

// Connect with default options.
client, err := figg.Connect("10.26.104.52:8119")
if err != nil {
	// handle err
}
```

Options can be provided, such as `WithReadBufLen`, described in `options.go`

### Subscribe
Subscribe to a topic to receive all messages published on that topic using
`Subscribe(name string, onMessage MessageCB, options ...TopicOption)`. Once
subscribed the SDK ensures no messages are dropped by recovering missed
messages while the client was disconnected.

`Subscribe` blocks until the server confirms the subscription is setup.

```go
err := client.Subscribe("foo", func(m figg.Message)) {
	fmt.Println("message: ", string(m.Data), "offset: ", m.Offset)
})
if err != nil {
	// handle err
}
```

`figg.Message` is described in `message.go`. This contains both a `Data` field
containing the published data, and an `Offset` field which points to the
location of next message in the topic.

The SDK uses this offset to automatically recover missed messages when the
connection is dropped, but it can also be passed as an option to `Subscribe`
using `WithOffset` to continue from an old message (such as may persist the
offset of the last message received to resume later).

### Publish
Publish a message to topic `foo` using
`Publish(name string, data []byte, onACK func())`.
```go
client.Publish("foo", []byte("bar"), func() {
	fmt.Println("message acked")
})
```

To acheive high thoughput the SDK supports sending multiple message before
the first has been acknowledged (similar to TCP), though does have a limit on
the number of unacknowledged messages (configured with `WithWindowSize`). If
the clients connection drops all unacknowledged will be retried (in order).

If its important to wait for each message to be acknowledged before sending the
next a wrapper `PublishWaitForACK` can be used.

Note you do not need to be subscribed to publish a message to a topic.
