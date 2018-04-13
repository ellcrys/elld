package peer

import "fmt"

// DoGetAddr sends GetAddr message to peers.
// Does not continue of active peers is greater or equal to 1000
func (protoc *Inception) DoGetAddr() error {

	if !protoc.PM().NeedMorePeers() {
		return nil
	}

	activePeers := protoc.PM().GetActivePeers(0)

	fmt.Println("Need more", activePeers)
	return nil
}
