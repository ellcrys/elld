package main

import (
	"flag"
	"log"

	"github.com/ellcrys/garagecoin/garagecoin"
)

const help = `Entry point to trigger the garagecoin service`

func main() {

	flag.Usage = func() {
		log.Println(help)
		flag.PrintDefaults()
	}

	destPeer := flag.String("d", "", "destination peer address")

	proxyPort := flag.Int("p", 9900, "proxy port")

	listenPort := flag.Int("l", 12000, "listen port")

	linkNodePeer := flag.String("link", "", "Node peer address")

	flag.Parse()

	if *linkNodePeer != "" {
		garagecoin.Link(proxyPort, linkNodePeer)
	} else {
		garagecoin.Run(destPeer, proxyPort, listenPort)
	}

}
