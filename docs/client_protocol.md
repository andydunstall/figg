# Client Protocol
This describes how the Wombat client interacts with the backend.

Note I'm using terminology:
* Client: The SDK implementation,
* Server: The Wombat cluster,
* User: The developer using the client SDK.

## Transports
WebSocket is the only supported transport. This was chosen as it is TCP based
and works in a browser.

## Connection
When the client is created it attempts to connect to the server.

WebSocket connects at URL `ws://{addr}/v1/ws`.

The address should be a load balancer routing the request to a random node
in the nearest region.

### Heartbeats
Once connected the client must send a `PING` message every 5 seconds, which the
server responds with a `PONG`. This `PING` includes a timestamp so the client
can monitor the latency between itself and the server.

If the client doesn't get a `PONG` within 5 seconds (or a configurable timeout)
it should attempt to reconnect.

Also if the server doesn't get a `PING` within 10 seconds of the initial
connection request or the last `PING` it assumes the client has disconnected and
closes the connection.

### Reconnect
If the connection drops clients reconnect. Retries use exponential backoff,
calculated as `100ms * min(2**num_attempts, 100)`.

## Topic
A user opens a topic with `wombat.Topic(name)`. This subscribes to that topic
so then can received messages, such as with the `topic.MessagesCh()` channel
in Go, and publish messages with `topic.Publish(message)`.

### Attach
When the user opens a topic, the client attaches to this topic by sending an
`ATTACH` message. Once the server has attached to the topic it will respond with
a `ATTACHED` response. Once attached the server forwrads all messages on the
topic to the client.

### Publish
Published messages are sent with the `PUBLISH` type. This includes the message
itself, which is just an opaque sequence of bytes, the topic name, and a
sequence number.

This sequence number must be incremented for each message sent by the client
(using the same counter for all topics). Though resends must use the same
sequence number as the initial attempt.

The sequence number is used to acknowledge messages. Once the server has
processed a message it will respond with a `ACK` message including the
highest sequence number it has acknowledged.

If messages are not acknowledged within 5 seconds the client should retry.

### Subscribe
Once attached the server will send all messages received on the topic. These
processed messages are assigned a unique serial.

The received messages have type `MESSAGE` and include the topic name, the
serial and the message payload.

### Reattach
If the clients connect drops it will reconnect to the server. It will then need
to re-attach all its topics.

To avoid dropped or duplicate messages, when re-attaching the client must
include the serial of the last message received on the topic. The server can
then send all messages since this serial that the client has missed.

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

**MESSAGE**
* Type: `5`
* Name: `message`
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
