package misttriggers

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItCallsFunctions(t *testing.T) {
	broker := NewTriggerBroker()
	mu := sync.Mutex{}
	calls := 0
	increment := func(ctx context.Context, payload *StreamBufferPayload) error {
		require.Equal(t, payload.StreamName, "TestStream")
		mu.Lock()
		defer mu.Unlock()
		calls += 1
		return fmt.Errorf("something went wrong")
	}
	broker.OnStreamBuffer(increment)
	broker.OnStreamBuffer(increment)
	broker.OnStreamBuffer(increment)
	broker.TriggerStreamBuffer(context.Background(), &StreamBufferPayload{StreamName: "TestStream"})
	require.Equal(t, calls, 3)
}
