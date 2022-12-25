# Server

The server runs two services:
* Messaging service: Which provides the core pub/sub functionality,
* Admin service: Which provides endpoints for admin debugging.

The admin services exposes [pprof](https://pkg.go.dev/net/http/pprof) endpoints
for debugging.

## Config
All configuration is passed via the command line. See `figg-server -h` for a
full list of options.
