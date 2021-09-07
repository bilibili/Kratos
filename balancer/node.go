package balancer

import (
	"context"
	"time"
)

// DoneInfo is callback info when RPC invoke done
type DoneInfo struct {
	// Response Error
	Err error
	// Response Metadata
	ReplyMeta Metadata

	// BytesSent indicates if any bytes have been sent to the server.
	BytesSent bool
	// BytesReceived indicates if any byte has been received from the server.
	BytesReceived bool
}

// Metadata is Node Meatadata
type Metadata interface {
	Get(key string) string
}

// Done is callback function when RPC invoke done
type Done func(ctx context.Context, di DoneInfo)

// Node is node interface
type Node interface {
	// address is unique under the same service
	Address() string

	// pick a node
	Pick() Done

	// runtime calcuated weight
	Weight() float64

	// time elapsed since the latest pick
	PickElapsed() time.Duration

	// Metadata is the kv pair metadata associated with the service instance.
	Metadata() Metadata
}

// NodeBuilder is node builder
type NodeBuilder interface {
	Build(addr string, initWeight float64, metadata Metadata) Node
}
