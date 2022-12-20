# Figg Go SDK

## Usage
### Connect
Users start by connecting to the Figg node.

Note though the user waits for the initial connection to succeed, if the
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

### Subscribe
Subscribe to a topic. Once subscribed the client ensure no messages are dropped,
even if the clients connection is dropped.

```go
// Subscribe. Blocks until the server confirms the subscription is setup.
err := client.Subscribe("foo", func(m figg.Message)) {
	fmt.Println("message: ", string(m.Data), "offset: ", m.Offset)
})
if err != nil {
	// handle err
}
```

Messages received by the subscriber include a `Data` field containing the
published data, and an `Offset` field which can be used to subscribe from a new
client without missing messages. Such as may persist the offset and subscriber
later with no dropped messages while disconnected.

Note the format of the `Offset` field is defined by the server so should only
use an offset of a received message rather than calculating it yourself. Such
as some bits are reserved for using as flags so its not nessesarily sequential.

```go
// Subscribe from offset.
err := client.Subscribe("bar", func(m figg.Message), WithOffset(offset)) {
	fmt.Println("message: ", string(m.Data), "offset: ", m.Offset)
})
if err != nil {
	// handle err
}
```

### Publish
Publish a message to topic `foo`:

```go
client.Publish("foo", []byte("bar"))
```

This will block until the server has acknowledged the published mesage.
