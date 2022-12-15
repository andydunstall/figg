package sdk

import (
	"testing"

	figg "github.com/andydunstall/figg/sdkv2"
	"github.com/stretchr/testify/assert"
)

func TestConnect_ServerUnreachable(t *testing.T) {
	_, err := figg.Connect("1.2.3.4:8119")
	assert.Error(t, err)
}
