# SDK

## Usage

**Connect**
Connects to a figg cluster. `Addr` is expected to be a load balancer that
distributes connections evenly among figg nodes in the nearest region.

```go
config := &Config{
	Addr: "my-figg-lb.com",
}
client := figg.NewFigg(config)
```

### Publish
Publishes a message to topic `foo`.

```go
if err := client.Publish("foo", []byte("bar")); err != nil {
	// ...
}
```

### Subscribe
Subscribes to topic `foo`.

```go
sub := client.Subscribe("foo", func(topic string, m []byte) {
  fmt.Println("received message", string(m))
})

client.Unsubscribe("foo", sub)
```
