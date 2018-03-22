package garagecoin

import (
	"bufio"
	"context"
	"fmt"
	host "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	"io"
	"log"
	"net/http"
	"strings"

	ps "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"

	ghost "github.com/ellcrys/garagecoin/garagecoin/ghost"
	gproxy "github.com/ellcrys/garagecoin/garagecoin/proxy"
)

type GossipService struct {
	host      host.Host
	dest      peer.ID
	proxyAddr ma.Multiaddr
}

const Protocol = "gragecoin/0.0.1"

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

// Run the garagecoin node server and proxy server
func Run(destPeer *string, proxyPort *int, listenPort *int) {

	if *destPeer != "" {

		host := ghost.NewHost(*listenPort + 1)
		getPeerID := addAddrToPeerstore(host, *destPeer)

		proxyAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", *proxyPort))

		if err != nil {
			log.Fatalln(err)
		}
		_ = proxyAddr
		proxy := gproxy.NewProxyService(host, proxyAddr, getPeerID)
		proxy.Serve()
	} else {

		host := ghost.NewHost(*listenPort)

		_ = gproxy.NewProxyService(host, nil, "")

		<-make(chan struct{})

	}

}

//Link a node peer address with this host
func Link(listenPort *int, addr *string) {
	ipfsadrr, err := ma.NewMultiaddr(*addr)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsadrr.ValueForProtocol(ma.P_IPFS)

	if err != nil {
		log.Fatalln(err)
	}

	//determine the peer id of the given nodePeer addr
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))

	targetAddr := ipfsadrr.Decapsulate(targetPeerAddr)

	//store this peer address
	host := ghost.NewHost(*listenPort)

	host.Peerstore().AddAddr(peerid, targetAddr, ps.PermanentAddrTTL)

	log.Println("Node Peer Connected: ", peerid)

	//_ = gproxy.NewProxyService(host, nil, "")

	//Announce my presence
	gs := NewGossipService(host, targetAddr, peerid)

	gs.Gossip()

	<-make(chan struct{})
}

//Gossip Announce to linked peer of my presence
func (gs *GossipService) Gossip() {

	_, serveArgs, _ := manet.DialArgs(gs.proxyAddr)

	log.Println("gossiping and listening to peer on ", serveArgs)

	if gs.dest != "" {
		http.ListenAndServe(serveArgs, gs)
	}
}

func (gs *GossipService) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	fmt.Printf("proxying request for %s to peer %s\n", req.URL, gs.dest.Pretty())

	//send request to remote host
	stream, err := gs.host.NewStream(context.Background(), gs.dest, Protocol)
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//remember to close our stream
	defer stream.Close()

	//Write request to the stream
	req.Write(stream)

	//retrieve response from the destination peer
	buff := bufio.NewReader(stream)

	resp, err := http.ReadResponse(buff, req)

	if err != nil {
		log.Println(err)
		stream.Close()
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//copy header
	for k, v := range resp.Header {
		for _, s := range v {
			writer.Header().Add(k, s)
		}
	}

	//Write the status
	writer.WriteHeader(resp.StatusCode)

	//copy the body
	io.Copy(writer, resp.Body)
	resp.Body.Close()
}

func NewGossipService(h host.Host, proxyAddr ma.Multiaddr, dest peer.ID) *GossipService {

	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.

	streamHandler := func(stream inet.Stream) {
		//remember to close the stream when we are done
		defer stream.Close()

		//Create a buffer reader to read the incoming request
		buff := bufio.NewReader(stream)

		//read the http request
		req, err := http.ReadRequest(buff)

		if err != nil {
			stream.Reset()
			log.Println(err)
			return
		}

		//Get the address of gossiper
		gossipper := req.Header.Get("myaddr")

		gsMaddr, err := ma.NewMultiaddr(gossipper)
		if err != nil {
			log.Println(err)
		}

		pid, err := gsMaddr.ValueForProtocol(ma.P_IPFS)

		if err != nil {
			log.Println(err)
		}

		//determine the peer id of the given nodePeer addr
		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Println(err)
		}
		h.Peerstore().AddAddr(peerid, gsMaddr, ps.PermanentAddrTTL)
		log.Println("Received and stored peer gossiper =>", peerid)

		defer req.Body.Close()

		//Set URL scheme to http

		hp := strings.Split(req.Host, ":")
		if len(hp) > 1 && hp[1] == "443" {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}

		//set request URL host
		req.URL.Host = req.Host

		//set new out going request
		outreq := new(http.Request)
		*outreq = *req

		outreq.Header.Add("myaddr", proxyAddr.String())

		//we now make the request
		log.Printf("Making request to %s\n", req.URL)
		resp, err := http.DefaultTransport.RoundTrip(outreq)
		if err != nil {
			stream.Reset()
			log.Println(err)
			return
		}

		//write response we got for our request, back to the stream
		resp.Write(stream)
	}

	h.SetStreamHandler(Protocol, streamHandler)

	log.Println("node-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", a, peer.IDB58Encode(h.ID()))
	}
	return &GossipService{
		host:      h,
		dest:      dest,
		proxyAddr: proxyAddr,
	}
}
