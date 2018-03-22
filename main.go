package main

import (
	"flag"
	"log"

	garagecoin "github.com/ellcrys/garagecoin/garagecoin"
)

const help = ``

func main() {

	flag.Usage = func() {
		log.Println(help)
		flag.PrintDefaults()
	}

	destPeer := flag.String("d", "", "destination peer address")

	proxy_port := flag.Int("p", 9900, "proxy port")

	listenPort := flag.Int("l", 12000, "listen port")

	flag.Parse()

	gcoin := new(garagecoin.GarageCoin)

	gcoin.Run(destPeer, proxy_port, listenPort)

}
