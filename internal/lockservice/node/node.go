package node

// Node describes the elements of a node of the distributed lockservice.
// This
type Node interface {
	// Start starts up the node by initialising the necessary data/
	Start() error
}
