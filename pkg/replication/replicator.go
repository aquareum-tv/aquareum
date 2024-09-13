package replication

import "context"

type Replicator interface {
	NewSegment(context.Context, []byte)
}
