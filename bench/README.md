# Bench
Bench is a tool for running benchmarks against Figg.

This tests a few metrics listed below, and can be configured to use different
message sizes, number of messages, and number of concurrent publishers,
subscribers and resumers.

A lot of this is from [natscli](https://github.com/nats-io/natscli).

## Metrics
* Publish: Measures the time it takes for N messages to be published and
acknowledged,
* Subscribe: Measures the time between the first message the the Nth message
is received (since we don't care that much about the time it takes to attach
given this should be rare),
* Resume: Subscribers from an offset of 0 **after** all messages have been
published, so this only receives messages from history.

## Example
```bash
$ figg-bench
starting benchmark [msgs=1,000,000 msg-size=128B topic=bench-43517 addr=127.0.0.1:8119 publishers=1 subscribers=1 resumers=1]
Sub stats: 835,068 msgs/sec ~ 101.94 MB/sec
Pub stats: 835,103 msgs/sec ~ 101.94 MB/sec
Resume stats: 319,738 msgs/sec ~ 39.03 MB/sec
```
