# Client Protocol

## Connection
Clients connect to Figg over TCP. Since Figg currently only supports a single
node, the address of that node is passed to the client.

### Ping/Pong
*TODO*

### Reconnect
If the client detacts the connection has dropped, either by pings timing our
or `read` returning an error, it will automatically reconnect. The client
currently retries forever, using exponential backoff by to avoid overloading
the server (though the user can provide a custom backoff strategy).

A user can register to be notified about connection state events, such as
disconnected and connected.

On reconnecting the client will handle re-sending any required messages as
described below.

## Topic
Clients publish and subscribe to topics. Message are is just an opaque blob
of bytes.

Each message published to a topic is assigned a `uint64` offset in the topic.
This offset points to the next message in the topic. This is used by the client
to subscribe from the offset of the last received message, which will resume
from the next message in the topic rather than the most recent. See
[topics.md](./topics.md) for details.

### Attachment
To subscribe to messages published to a topic the client sends an `ATTACH`
request. This may include an offset field containing the offset of an old
message message received (typically the last message received to ensure
continuity).

The server responds with an `ATTACHED` message containing the offset the
subscription has started from. If no offset was included in `ATTACH` this will
be the offset of the most recent message.

Note the offset may not match the requested offset. This happens if the
requested offset is expired (where the expiry is configurable on the server).
In this case the server uses the offset of the oldest on the topic.

#### Messages
Once attached the client receives messages from the topic since the attachment
offset. This will stream historical messages (when the offset is less than the
latest message) as fast as it can, then new messages will be sent at they are
published.

Messages are received as `DATA` messages, which contains the topic name,
offset and the published data.

#### Detach
To unsubscribe the client sends a `DETACH` request. The server responds with
a `DETACHED` message so the client can clear any state and stop retrying on
reconnect.

#### Reconnect
If the client disconnects, once it reconnects it tries to recover from where it
left off, maintaining message continuity.

On reconnect any attaching attachments, where an `ATTACH` message has been sent
but not recieved an `ATTACHED` response, will be resent (with the same offset if
given).

For attached topics, the client tracks the offset of the last message recieved
(or the offset from `ATTACHED` if no messages have been received). When the
client reconnects it sends an `ATTACH` with this tracked offset so it can
resume from where it left off.

Any detaching topics, where they have sent `DETACH` but not received `DETACHED`,
will also resend the `DETACH` message (unless it has since been re-attached).

## Publish
The client publishes messages by sending `PUBLISH` messages.

Note the client doesn't need to be attached to publish. They only attach to
subscribe.

To guarantee delivery, the client must retry any messages that have not been
received by the server.

To acheive this, the client assigns each published message a unique sequence
number (which must be greater than all previous published messages). It includes
the sequence number in the `PUBLISH` message, sends this to the server then
stores all unacknowledged messages.

Once the server has processed the message it responds with an `ACK` that contains
the sequence number of the last message processed. When the client receives this
`ACK` it can clear any stored messages with a sequence number equal to or less
than this number.

If the clients connection drops, when it reconnects it resends all unacknowleged
messages (in order).

Note the connection could drop after the server processes the publish but before
it sends the ACK, which would cause the client to resend the message leading
to duplicates. This means the service provides at least once delivery, rather
than exactly once delivery, since guaranteeing exactly once delivery would add
so much overhead the service would be too slow.

In the future should probably limit the number of pending messages waiting to
be acknowledged but for now its unbounded.

## Protocol
The Figg protocol uses a simple binary protocol to encode messages.

Each messages starts with an 8 byte header containing:
* Message type: `uint16`
  * Used for routing the message to the appropriate handler,
* Protocol version: `uint16`
  * Currently `1`
* Payload size: `uint32`
  * Size of the messge payload in bytes

The payloads contain zero or more fields. Variable size fields are encoded
as `[]byte` and prefixed with a `uint32` containing its size.

Integers are encoded in network byte order.

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

#### DATA
* Message type: `7`
* Direction: Server -> Client
* Fields
  * `topic` ([]byte)
  * `offset` (uint64)
  * `data` ([]byte)
