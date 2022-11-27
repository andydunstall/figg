# Topics
**WIP**

This describes the design of Figg topics within the service. Here subscribers
refers to server side subscribers that send messages to the client over the
client connection.

Topics are designed to support:
* Retention: Storing messages for a configurable window on disk (such as a
week),
* Resume from offset: Support subscribing from a specific message offset, such
as when a client disconnects they can resume from that previous offset,
* Resume from timestamp: Support subscribing from a timestamp, such as a client
may subscribe starting at messages 24 hours ago.

The design of topics is based on [moss](https://github.com/couchbase/moss/blob/master/DESIGN.md).

## Commit Log
Keep in memory repr same as commit log

## Subscribers
There are two types of subscriber:
* Resuming subscriber: A subscriber that is resuming from some offset in the
past, iterating though the topic history until it is up to date,
* Live subscriber: A subscriber that has all messages on the topic so is waiting
for new messages.

## Live Subscribers
Each topic maintains the set of live subscribers. Whenever a message arrives
at a topic the first thing it does is send the the live subscribers.

To avoid this blocking when interating though the list of subscribers, each
connection maintains an output buffer (of finite size) to be sent to the client.
So 'sending' a message to a subscriber really just means encode the protocol
message and append it to the connections output buffer (which the connection
goroutine will send itself).

If the output buffer becomes full the connection is closed. This will cause
the client to reconnect and resume from an offset, becoming a resuming
subscriber. This only way this should happen is if the client cannot keep up
with the rate of messages, though Figg isn't really designed for high thoughput
so this should be rare.

For live subscribers this is very similer to what Redis pub/sub does.

## Resuming Subscribers
Wheras live subscribers are blocked waiting for more messages from the topics
they are subscribed to, resuming subscribers have a backlog of messages to send
to the client.



Each has its own goroutine
* get a segment from the commit log
* send messages from that segment to the client
* get a new segment


Unlike live subscribers



When a new mess






are not sent to the subscriber,
* Live subscriber: The subscriber is up to date with messages in the topic
so any new messages are sent to that subscriber.



## Live Subscribers
Live subsc

Live subscribers are relatively up to date with the current topic













The aims of the topics are:
* Low latency: Tries to be low latency both subscribing to new incoming (live)
messages and historical messages,
* Retention: Stores messages for a configurable window which is expected to be
large (such as a week)


## Topic Layout
The topic 

Each topic maintains two slices of bytes (`[]byte`)


Topics are split into chunks







* Low latency when subscribing to up to date messages








Split into two types of subscriber
* live subscribers - subscriptions up to date so expect to get all incoming
messages immediately
* non-live subscribers - subscriptions who are still iterating though the historic
messages to catch up - once caught up will become live subscribers

## Publishing
1. add to output buffers of all live subscribers (does this need a copy or will a
pointer do?)
2. add to topic itself (topic.Add(message))
  * this persists to the commit log

note the purpose of the commit log isn't fault tolerance - just for retention
when the historical messages can't fit in memory

build up a single slice as described in the talk - fewer objects in memory for
the GC and easy to write to commit log in batches - also keep current chunk in
memory

## Subscribing

if not subscribing with a particular offset just added straight to the
live subscribers

if getting an offset - will iterate though the log (the latest of which may
be in memory) sending to the client

fetches from the commit log in chunks (configurable - maybe 1MB to start with?) -
then keeps that chunk in memory (cached by the topic) - and the subscriber iterates
though sending to the client (in the clients own goroutine) - once sent will
get the next chunk and so on
once up to date will add to the live subscribers set (this will need to be
atomic to avoid missing messages)
