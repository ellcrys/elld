package garagecoin

import (
	"fmt"
	host "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	"log"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	ghost "github.com/ellcrys/garagecoin/garagecoin/ghost"
	gproxy "github.com/ellcrys/garagecoin/garagecoin/proxy"

	ps "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
)

type GarageCoin struct{}

// addAddrToPeerstore parses a peer multiaddress and adds
// it to the given host's peerstore, so it knows how to
// contact it. It returns the peer ID of the remote peer.
func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	ipfsadrr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsadrr.ValueForProtocol(ma.P_IPFS)

	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))

	targetAddr := ipfsadrr.Decapsulate(targetPeerAddr)

	h.Peerstore().AddAddr(peerid, targetAddr, ps.PermanentAddrTTL)

	return peerid
}

const help = `
	Usage: 
	Start remote peer first with:   ./garaagecoin ==> This will print the remote peer multiaddress to connect to
	Start the local peer with: ./garagecoin -d <remote proxy address>

`

func (g *GarageCoin) new() *GarageCoin {

	return &GarageCoin{}
}

func (g *GarageCoin) Run(destPeer *string, proxy_port *int, listenPort *int) {

	if *destPeer != "" {
		ghost := new(ghost.Host)
		host := ghost.NewHost(*listenPort + 1)

		getPeerID := addAddrToPeerstore(host, *destPeer)
		log.Println(getPeerID)
		proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", proxy_port))

		if err != nil {
			log.Fatalln(err)
		}
		//proxy := ghost.NewProxyService(host, proxyAddr, destPeerID)

		gproxy := new(gproxy.ProxyService)
		proxy := gproxy.NewProxyService(host, proxyAddr, getPeerID)
		proxy.Serve()
	} else {
		ghost := new(ghost.Host)
		host := ghost.NewHost(*listenPort)

		gproxy := new(gproxy.ProxyService)
		_ = gproxy.NewProxyService(host, nil, "")

		<-make(chan struct{})

	}

}
