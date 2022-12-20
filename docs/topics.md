# Topics
This describes the design of Figg topics within the server, so here subscriber
refers to a server side structure (that forwards messages to the client) rather
than a client subscriber.

Topics store messages for a configurable on disk in a commit log.

See [`service/pkg/topic`](../service/pkg/topic).

## Commit Log
The commit log is split into roughly 4MB segments. Each segment is assigned an
offset in the commit log. The most recent segment is kept in memory so publishes
are fast, then once its full it is persisted to disk in the background.

See [`service/pkg/commitlog`](../service/pkg/commitlog).

#### `Append(message)`
1. Looks up the most recent segment, which will always stored in-memory as a
`[]byte` slice,
2. Appends the message to the segment, prefixed by a 32 bit message size,
3. If the segment is full a new in-memory is created (becoming the most recent
segment), and a goroutine is spun up to persist the full segment (to avoid
append blocking)

#### `Persist(segment)`
To persist an in-memory segment, the format on disk is the same as the in-memory
segment so this can just be written to a new file. Each segment has its own
file to make retention easy (when a segment is expired it can be deleted)

#### `Lookup(offset)`
1. Looks up the segment that contains the offset (in an in-memory structure
mapping offsets to segments),
2. If not found returns,
3. Otherwise looks up in that segment, which is either in-memory or on disk.

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
