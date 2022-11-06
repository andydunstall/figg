# Client Protocol
This describes how the Wombat SDK interacts with the backend.

## Transports
WebSocket is the only supported transport. This was chosen as it is TCP based
and works in a browser.

## Connection
When the client is created it attempts to connect to the server.

WebSocket connects at URL `ws://{addr}/v1/ws`.

The address should be a load balancer routing the request to a random node
in the nearest region.

### Heartbeats
Once connected the client must send a `PING` message every 5 seconds
(configurable), which the server responds with a `PONG`. This `PING` includes a
timestamp so the client can monitor the latency between itself and the server.

If the client doesn't get a `PONG` by the time it next sends a `PING` it assumes
it has disconnected so reconnects.

Also if the server doesn't get a `PING` within 10 seconds of the initial
connection request or the last `PING` it assumes the client has disconnected and
closes the connection.

### Reconnect
If the connection drops clients reconnect. Retries use exponential backoff,
calculated as `100ms * min(2**num_attempts, 100)`.

## Topic
### Publish
A user publishes a message by calling `client.Publish(topic, message)`. The
SDK will send a protocol message with the `PUBLISH` type. This includes the
users message payload, which is just an opaque seqeuence of bytes, the topic
name, and a sequence number.

This sequence number is incremented for each message sent by the client
(using the same counter for each connection). Though resends use the same
sequence number as the initial attempt.

The sequence number is used to acknowledge messages. Once the server has
processed a message it will respond with a `ACK` message including the
highest sequence number it has acknowledged.

If messages are not acknowledged within 2 seconds client retries.

### Subscribe
A user subscribes to a channel with `client.Subscribe(topic)`. The SDK will
attach to this topic by sending an `ATTACH` message with the topic name. Once
the server has attached (by connecting to the coordinator for that topic) it
will respond with an `ATTACHED` response. Once attached the server forwards
all messages on the topic to the client in `PAYLOAD` messages.

This `PAYLOAD` message contains the topic, the user payload and a serial
used to uniquely identify the message on the topic. Note this serial is
different from the sequence number used when publishing, and is used to
reattach if the connection drops (see below).

The SDK then calls the user provided subscriber with the payload containing
a message published on the topic.

**Reattach**

When the connection to the server is dropped, we could potentially miss
messages published on the topic.

To handle this, when a new connection is established, the SDK reattaches to
all previously attached topics and includes the serial of the last message
on that topic. The server then starts attaches from that point so sends any
messages we may have missed (as long as that message is still retained by the
server). The client detects whether we have missed messages when the serial
in the `ATTACHED` response doesn't match the serial it requested in `ATTACH`.

## Protocol
The Wombat protocol uses msgpack to encode all messages. Each protocol message
has a `uint16` type and format:
```
{
  type: <type: uint16>,
  <name: string>: {
    (fields)
  }
}
```

Such as a `ATTACHED` message would have format:
```
{
  type: 2,
  attached: {
    topic: <topic: string>,
    message: <message: uint8[]>,
    sequence_number: <sequence_number: uint32>,
  }
}
```

### Types
**ATTACH**
* Type: `1`
* Name: `attach`
* Direction: Client -> Server
* Fields
  * `topic` (string)
  * `serial` (string)

**ATTACHED**
* Type: `2`
* Name: `attached`
* Direction: Server -> Client
* Fields
  * `topic` (string)
  * `serial` (string)

**PUBLISH**
* Type: `3`
* Name: `publish`
* Direction: Client -> Server
* Fields
  * `topic` (string)
  * `message` (uint8[])
  * `sequence_number` (uint32)

**ACK**
* Type: `4`
* Name: `ack`
* Direction: Server -> Client
* Fields
  * `topic` (string)
  * `sequence_number` (uint32)

**PAYLOAD**
* Type: `5`
* Name: `payload`
* Direction: Server -> Client
* Fields
  * `topic` (string)
  * `serial` (string)
  * `message` (uint8[])

**PING**
* Type: `6`
* Name: `ping`
* Direction: Client -> Server
* Fields
  * `timestamp` (uint64)

**PONG**
* Type: `7`
* Name: `pong`
* Direction: Server -> Client
* Fields
  * `timestamp` (uint64)
