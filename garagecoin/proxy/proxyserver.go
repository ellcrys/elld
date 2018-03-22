package gproxy

import (
	"bufio"
	"context"
	"fmt"
	host "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	manet "gx/ipfs/QmRK2LxanhK2gZq6k6R7vk5ZoYZk8ULSSTB7FzDsMUX6CB/go-multiaddr-net"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	"io"
	"log"
	"net/http"
)

//Protocol defines the libp2p protocol
const Protocol = "gragecoin/0.0.1"

//ProxyService struct
type ProxyService struct {
	host      host.Host
	dest      peer.ID
	proxyAddr ma.Multiaddr
}

//NewProxyService creates a new proxy service
func (p *ProxyService) NewProxyService(h host.Host, proxyAddr ma.Multiaddr, dest peer.ID) *ProxyService {

	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	h.SetStreamHandler(Protocol, streamHandler)
	log.Println("Proxy server is ready")
	log.Println("libp2p-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", a, peer.IDB58Encode(h.ID()))
	}
	return &ProxyService{
		host:      h,
		dest:      dest,
		proxyAddr: proxyAddr,
	}
}

func streamHandler(stream inet.Stream) {

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

	defer req.Body.Close()

	//Set URL scheme to http
	req.URL.Scheme = "http"

	//set request URL host
	req.URL.Host = req.Host

	//set new out goiung request
	outreq := new(http.Request)
	*outreq = *req

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

//Serve launch the proxy service to be listened to
func (p *ProxyService) Serve() {
	_, serveArgs, _ := manet.DialArgs(p.proxyAddr)

	log.Println("proxy listening on ", serveArgs)

	if p.dest != "" {
		http.ListenAndServe(serveArgs, p)
	}
}

//
func (p *ProxyService) ServeHTTP(writer http.ResponseWriter, req *http.Request) {

	fmt.Printf("proxying request for %s to peer %s\n", req.URL, p.dest.Pretty())

	//send request to remote host
	stream, err := p.host.NewStream(context.Background(), p.dest, Protocol)
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
