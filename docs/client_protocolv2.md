# Client Protocol

## Connection
Clients connect to Figg over TCP. Since Figg currently only supports a single
node, the address of that node is passed to the client.

### Ping/Pong

### Reconnect
If the client detects the connection has dropped it will automatically
reconnect, with exponential backoff to avoid overloading the server.

A user callback can be provided to be notified about connection state events,
such as disconnected and connected

## Attachment
When subscribing to a topic the client sends an `ATTACH` message containing
the name of the topic to subscribe to. This may include an offset of the last
message the client received, which the server will subscribe to the next message
after this.

Once the server sets up the subscription it responds with an `ATTACHED` message
containing the offset of the most recent message that will not be included in
the subscription. The client can detect a message discontinuity if this offset
doesn't match the requested offset, which occurs if the requested offset has
expired.

### Re-attach
The client tracks the offset of the most recent message received on the topic,
or the offset contained in `ATTACHED` if not yet received any messages.

If the connection drops, the client automatically reconnects. When a new
connection succeeds an `ATTACH` message for all attached and attaching topics
is sent containing this tracked offset to recover any messages missed while
disconnected.

### Detach
To unsubscribe the client sends a `DETACH` request with the topic name. This
is retried on each reconnect until a `DETACHED` response is received.

Note if the user subscribes to the topic again before receiving `DETACHED` it
stops retrying.

## Publish
The client publishes messages with the `PUBLISH` message type, which contains
the topic name and message data.

Since the connection to the server can drop at any time, publishes must be
acknowledged at the application level.

The client gives each published message a unique sequence number, that is
incremented for each message published. Note the same sequence number counter
is used for all publishes on the connection (which may have multiple topics).

Once the server processes the published message it responds with an `ACK`
containing the sequence number of the published message.

The client must resend all publishes that have not been acknowledged when it
reconnects, sending in the same order as the original publishes. This means
it must store all unacknowledged messages in order, then resend once the
client reconnects.

Once the client receives an `ACK` it can discard all messages with a sequence
number equal to or less than that seqence number.

Note the connection could drop after the server processes the publish but before
it sends the ACK, which would cause the client to resend the message leading
to duplicates. This means the service provides at least once delivery, rather
than exactly once delivery, since guaranteeing exactly once delivery would add
so much overhead the service would be too slow.

In the future should probably limit the number of pending messages waiting to
be acknowledged but for now its unbounded.

## Protocol
The Figg protocol uses a custom binary protocol to encode messages.

Each messages starts with an 8 byte header containing:
* Message type: `uint16`
  * Used for routing the message to the appropriate handler,
* Protocol version: `uint16`
  * Currently `1`
* Payload size: `uint32`
  * Size of the messge payload in bytes

The payloads may include zero or more fields. Variable size fields are encoded
as `[]byte` and prefixed with a `uint32` containing its size.

Integers and encoded in network byte order.

### Messages
#### ATTACH
* Message type: `1`
* Direction: Client -> Server
* Fields
  * `flags` `uint16`
    * Bit 1: If `1` subscribes from a particular offset given in the payload,
otherwise subscribes from the latest message on the topic (and the `offset`
field is unused)
  * `topic` ([]byte)
  * `offset` (uint64)

#### ATTACHED
* Type: `2`
* Name: `attached`
* Direction: Server -> Client
* Fields
  * `topic` (string)
  * `offset` (uint64)

#### DETACH
* Message type: `3`
* Direction: Client -> Server
* Fields
  * `topic` ([]byte)

#### DETACHED
* Message type: `4`
* Direction: Server -> Client
* Fields
  * `topic` ([]byte)

#### PUBLISH
* Message type: `5`
* Direction: Client -> Server
* Fields
  * `topic` ([]byte)
  * `seq_num` (uint64)
  * `data` ([]byte)

#### ACK
* Message type: `6`
* Direction: Server -> Client
* Fields
  * `seq_num` (uint64)
