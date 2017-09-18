package main

import (
	"context"
	"fmt"
	"os"

	peer "github.com/libp2p/go-libp2p-peer"
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

	QmAlice, err := peer.IDFromString("QmAlice")
	if err != nil {
		fatal(err)
	}
	QmBob, err := peer.IDFromString("QmBob")
	if err != nil {
		fatal(err)
	}

	s := swarm.NewBlankSwarm(context.Background(), QmAlice, nil, nil)
	s.AddPacketTransport(udpt.NewUDPTransport())

	// Add an address to start listening on
	err = s.AddListenAddr(laddr)
	if err != nil {
		fatal(err)
	}

	// Conn as argument, for WriteMsg()?
	s.SetMsgHandler(func(msg []byte, p peer.ID) {
		fmt.Printf("got message from %s: %s\n", p, string(msg))

		_, err = s.WriteMsg(msg, p)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	s.WriteMsg("hey bob", QmBob)

	// Wait forever
	<-make(chan bool)
}
