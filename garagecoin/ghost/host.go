package ghost

import (
	"context"
	"crypto/rand"
	"fmt"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
	"log"

	ps "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"

	swarm "gx/ipfs/QmSwZMWwFZSUpe5muU2xgTUwppH24KfMwdPXiwbEp2c6G5/go-libp2p-swarm"

	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

//Host struct
type Host struct {
	port      int
	basichost *bhost.BasicHost
	multiaddr *ma.Multiaddr
}

//func (h *Host) new(listenPort int) {}

//NewHost : Create a new host
func (h *Host) NewHost(listenPort int) *bhost.BasicHost {
	// Generate an identity keypair using go's cryptographic randomness source
	priv, pub, err := crypto.GenerateSecp256k1Key(rand.Reader)

	if err != nil {
		log.Fatalln(err)
	}

	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		log.Fatalln(err)
	}

	//Make the listen multi address
	listen, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort))
	if err != nil {
		log.Fatalln(err)
	}

	peerstore := ps.NewPeerstore()
	peerstore.AddPrivKey(pid, priv)
	peerstore.AddPubKey(pid, pub)

	// Make a context to govern the lifespan of the swarm
	ctx := context.Background()

	// Put all this together: create the swarm network
	netw, err := swarm.NewNetwork(ctx, []ma.Multiaddr{listen}, pid, peerstore, nil)
	if err != nil {
		log.Fatalln(err)
	}

	basichost := bhost.New(netw)

	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basichost.ID().Pretty()))

	if err != nil {
		log.Fatalln(err)
	}

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basichost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)

	log.Printf("I am %s\n", fullAddr)

	h.basichost = basichost
	h.port = listenPort
	h.multiaddr = &listen

	return basichost

}
