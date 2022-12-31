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
