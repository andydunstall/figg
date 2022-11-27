# Topics
**WIP**

This describes the design of Figg topics within the service. Here subscribers
refers to server side subscribers that attach to a topic and send messages to
the client.

Topics are designed to support:
* Retention: Storing messages for a configurable window on disk (such as a
week),
* Resume from offset: Support subscribing from a specific message offset, such
as when a client disconnects they can resume from that previous offset,
* Resume from timestamp: Support subscribing from a timestamp, such as a client
may subscribe starting at messages 24 hours ago.

## Commit Log
Currently this is implemented in memory with a slice of byte slices
(`[][]byte`).

Plan to change to:
* Replace a slice of byte slices with a single byte slice (to avoid lots of
small slices),
* Add an on disk commit log.

This is based on the design of [moss](https://github.com/couchbase/moss/blob/master/DESIGN.md).

## Subscribers
Subscribers can be in two states:
* Resuming: A subscriber that is resuming from some offset, iterating though
the topic history, ignoring all new messages on the topic until it is up to
date,
* Live: A subscriber that is up to date so is waiting for new messages on the
topic.

### Live
Live subscribers simply attach to the topic. When the topic receives a new
message, it just iterates though the attached subscribers and sends the
message to the attachment.

Note the attachment must not block since the topic syncrously iterates though
each. Currently this send operation:
1. Takes the connections lock (which there should be little contention),
2. Appends the message to the outgoing buffer,
3. Signals the send loop goroutine to wake up with a condition variable.

In the future can look at limitting this output buffer. Such as if it becomes
full (due to the client not reading from the socket fast enough causing
backpressure in the send loop), then revert to a resuming subscriber (or just
close the connection which has the same affect).

This case of a live subscriber, where the topic just iterates over the
subscribers appending to their output buffers is basically the same as Redis
pub/sub (although since Redis is single threaded it doesn't have locks and
condition variables).

### Resuming
Resuming subscribers have a backlog of messages to send to the client so must
iterate though the topics history. Once they are up to date (no more messages
left to send) then become live subscribers. This transision must be atomic to
avoid missing messages (such as if there is a publish between the subscriber
checking if it is up to date and registering with the topic) (such as with
`topic.RegisterIfLatest(offset))`.
