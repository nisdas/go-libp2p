package main

import (
	"context"
	"fmt"
	"os"

	swarm "github.com/libp2p/go-libp2p-swarm"
	udpt "github.com/libp2p/go-udp-transport"
	ma "github.com/multiformats/go-multiaddr"
)

func fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

// XXX unrewritten broken because jbenet/go-stream-muxer is out of date
// TODO move to libp2p org: go-stream-muxer go-smux-multistream go-smux-spdystream go-smux-yamux
// TODO multigram will live in BasicHost, for now it's only swarm

func main() {
	laddr, err := ma.NewMultiaddr("/ip4/0.0.0.0/udp/4737")
	if err != nil {
		fatal(err)
	}

	s := swarm.NewBlankSwarm(context.Background(), "Qmbob", nil, nil)
	s.AddPacketTransport(udpt.NewUDPTransport())

	// Add an address to start listening on
	err = s.AddListenAddr(laddr)
	if err != nil {
		fatal(err)
	}

	// Conn as argument, for WriteMsg()?
	s.SetMsgHandler(func(msg []byte, peerid string) {
		fmt.Printf("got message from %s: %s\n", peerid, string(msg))

		_, err = s.WriteMsg(msg, peerid)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	s.WriteMsg("hey bob", "Qmbob")

	// Wait forever
	<-make(chan bool)
}
