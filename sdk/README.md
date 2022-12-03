# SDK

## Usage

### Connect
Connects to a figg node.

```go
config := &Config{
	Addr: "mynode",
}
client := figg.NewFigg(config)
```

### Publish
Publishes a message to topic `foo`.

```go
client.Publish("foo", []byte("bar"))
```

### Subscribe
Subscribes to topic `foo`.

```go
sub := client.Subscribe("foo", func(topic string, m []byte) {
  fmt.Println("received message", string(m))
})
defer client.Unsubscribe("foo", sub)
```
