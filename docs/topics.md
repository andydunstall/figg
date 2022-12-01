# Topics
**WIP**

This describes the design of Figg topics within the server, so here subscriber
refers to a server side structure (that forwards messages to the client) rather
than a client subscriber.

Topics are designed to support:
* Retention: Storing messages for a configurable window on disk (such as a
week),
* Resume: Support subscribing from a specific message offset, such
as when a client disconnects they can resume from the offset of the last message
they received.

Topics support messages of upto 256KB.

## Commit Log
The commit log stores all messages in an append only log. The design of the
commit log is based on [Kafka](https://notes.stephenholiday.com/Kafka.pdf)
(described in section 3.1).

The log is split up into segments or roughly 1GB, which with a maximum message
size of 256KB holds at least 3000 messages per segment. Each segment has its
own file.

To support looking up a specific offset, the offsets of each segment are
maintained in an in memory structure (and flushed to disk). So when a particular
offset is required, the server looks up the segment containing that offset,
subtracts the segments offset, and uses the result as the offset within that
segment file.

Unlike Kafka, messages do not have to be flushed to disk before being sent to
clients (in fact messages are sent before being added to the commit log at all).
The plan is for fault tolerance to be aceived by replication instead. So if
a server crashes, when it comes up it can recover most of the commit log from
disk, then fetch any messages it missed (either due to not being flushed or
published while the server was down) from a replica.

Messages prefixed by their size (with a `uint32`) before being appended to the
segment.

So when a message is added to the commit log:
1. Checks if the previous segment is full, if so creates a new one and adds
the offset to the offset structure,
2. Appends the message size and message itself to the segment file.

Note currently having to load read messages into memory, since the Websocket
has a simple `Send(b []byte)` interface. Once move to a custom implementation
that allows writing to the TCP socket directly can use sendfile.

## Subscribers
Subscribers can be in two states:
* Resuming: A subscriber that is resuming from some offset, iterating though
the topic commit log, ignoring all new messages on the topic until it is up to
date,
* Live: A subscriber that is up to date so is waiting for new messages on the
topic.

### Live
Live subscribers simply attach to the topic. When the topic receives a new
message, it just iterates though the attached subscribers and sends the
message to the attachment (before adding to the commit log).

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
iterate though the topics commit log. Once they are up to date (no more messages
left to send) then become live subscribers. This transision must be atomic to
avoid missing messages (such as if there is a publish between the subscriber
checking if it is up to date and registering with the topic) (such as with
`topic.RegisterIfLatest(offset))`.
