# SDK

## Usage

**Connect**
Connects to a wombat cluster. `Addr` is expected to be a load balancer that
distributes connections evenly among wombat nodes in the nearest region.

```go
config := &Config{
	Addr: "my-wombat-lb.com",
}
client := wombat.NewWombat(config)
```

### Publish
Publishes a message to topic `foo`.

```go
if err := client.Publish("foo", []byte("bar")); err != nil {
	// ...
}
```

### Subscribe
Subscribes to topic `foo`. The subscriber is of type `MessageSubscriber`, which
it an interface with a member `NotifyMessage(m []byte)` which the user can
implement. An implementation exists for writing all messages to a channel,
`ChannelStateSubscriber` used here.

```go
sub := NewChannelMessageSubscriber()
topic := client.Subscribe("foo", sub)
m := <-sub
```
