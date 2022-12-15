# Figg Go SDK

## Usage
### Connect
Users start by connecting to the Figg node.

Note though the user waits for the initial connection to succeed, if the
connection drops the SDK will automatically reconnect.

```go
import (
	"github.com/dunstall/figg/sdk"
)

// Connect with default options.
client, err := figg.Connect("10.26.104.52:8119")
if err != nil {
	// handle err
}
```
