# TODO

## Connection
* Retry connection
  * Either if the initial connection fails or if the connection drops
  * Testing
    * Test connecting 5 seconds before the server starts
	* Test server responds with RST for 2 seconds and client still reconnects
	* Test server never responds to ping and SDK reconnects
* Add heartbeats
  * Test by adding latency so never receive PONG
