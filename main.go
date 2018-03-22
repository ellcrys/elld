package main

import (
	"fmt"

	"github.com/ellcrys/garagecoin/components"
	"github.com/ellcrys/garagecoin/protocols/inception"
)

func main() {

	log := components.NewLogger("/main")

	log.Infof("Garagecoin node started")

	// create the peer
	peer, err := components.NewPeer(4500, 10)
	if err != nil {
		log.Fatalf("failed to create peer")
	}

	log.Infof("=> Address: %s", peer.GetAddress())

	// set protocol version and handler
	peer.SetProtocolHandler(inception.NewInception("/inception/0.0.1"))
	fmt.Println(peer.GetAddress())

	// cause main thread to wait for peer
	peer.Wait()
}
