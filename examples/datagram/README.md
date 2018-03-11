# Datagram protocols with libp2p

This libp2p example is a form of readme-driven development for packet switching.
I'm using it to test out interfaces and conventions, and also to produce working code.

For now it's going to be the simplest possible echo server/client.


## things to check for streams vs. packets

- dialsync/dialbackoff
- dialer
- filters
- transports
- bwc
- protector
- addrutil.FilterUsableAddrs


## things that we've touched so far

- [x] go-multiaddr-net
- [ ] go-libp2p-transport
- [ ] go-udp-transport
- [ ] go-libp2p-swarm
- [ ] go-libp2p-packetswitch
- [ ] go-libp2p-host
- [ ] go-multigram
- [ ] go-wireguard
