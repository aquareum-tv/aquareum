package proc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var example = "INFO|MistController|24550|../subprojects/mistserver/lib/socket.cpp:2474|streamname|UDP bind success 10 on 127.0.0.1:4242 (IPv4)"

func TestMistLogParser(t *testing.T) {
	level, procName, pid, path, streamName, msg, err := ParseMistLog(example)
	assert.NoError(t, err)
	assert.Equal(t, level, "INFO")
	assert.Equal(t, procName, "MistController")
	assert.Equal(t, pid, "24550")
	assert.Equal(t, path, "../subprojects/mistserver/lib/socket.cpp:2474")
	assert.Equal(t, streamName, "streamname")
	assert.Equal(t, msg, "UDP bind success 10 on 127.0.0.1:4242 (IPv4)")
}
