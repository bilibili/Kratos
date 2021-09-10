package selector

import (
	"context"
	"sync"
	"time"
)

// Balancer is balancer interface
type Balancer interface {
	Pick(ctx context.Context, nodes []WeightedNode) (selected WeightedNode, done Done, err error)
}

// WeightedNode calculates scheduling weight in real time
type WeightedNode interface {
	Node

	// Weight is the runtime calculated weight
	Weight() float64

	// Pick the node
	Pick() Done

	// PickElapsed is time elapsed since the latest pick
	PickElapsed() time.Duration
}

// WeightedNodeBuilder is WeightedNode Builder
type WeightedNodeBuilder interface {
	Build(Node) WeightedNode
}

// Default is composite selector
type Default struct {
	NodeBuilder WeightedNodeBuilder
	Balancer    Balancer
	Filters     []Filter

	lk            sync.RWMutex
	weightedNodes []Node
}

// Select select one node
func (d *Default) Select(ctx context.Context, opts ...SelectOption) (selected Node, done Done, err error) {
	d.lk.RLock()
	weightedNodes := d.weightedNodes
	d.lk.RUnlock()
	// filter nodes
	for _, f := range d.Filters {
		weightedNodes = f(ctx, weightedNodes)
	}
	var options SelectOptions
	for _, o := range opts {
		o(&options)
	}
	for _, f := range options.Filters {
		weightedNodes = f(ctx, weightedNodes)
	}
	candidates := make([]WeightedNode, 0)
	for _, n := range weightedNodes {
		candidates = append(candidates, n.(WeightedNode))
	}
	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailable
	}
	return d.Balancer.Pick(ctx, candidates)
}

// Apply update nodes info
func (d *Default) Apply(nodes []Node) {
	weightedNodes := make([]Node, 0)
	for _, n := range nodes {
		weightedNodes = append(weightedNodes, d.NodeBuilder.Build(n))
	}
	d.lk.Lock()
	// TODO: Do not delete unchanged nodes
	d.weightedNodes = weightedNodes
	d.lk.Unlock()
}
