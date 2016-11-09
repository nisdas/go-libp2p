# Datagram protocols with libp2p

This libp2p example is a form of readme-driven development for packet switching.
I'm using it to test out interfaces and conventions, and also to produce working code.

For now it's going to be the simplest possible echo server/client.

## Additions

package p2phost
type MsgHandler func(peer.ID, protocol.ID, []byte)
type Host interface {
  MsgMux() *msmux.MultigramMuxer
  SetMsgHandler(protocol.ID, inet.MsgHandler)
  SetMsgHandlerMatch(protocol.ID, func(string) bool, MsgHandler)
  RemoveMsgHandler(protocol.ID)
  WriteMsg(context.Context, peer.ID, ...protocol.ID, []byte)
}

package p2pswarm
type MsgHandler func(peer.ID, []byte)
func (s *Swarm) AddPacketTransport(p2ptransport.PacketTransport) {}
func (s *Swarm) SetMsgHandler(MsgHandler) {}
func (s *Swarm) MsgHandler() MsgHandler {}

package p2piconn
type PacketConn interface {}

package p2pnet

package p2pstream
func NoOpMsgHandler(msg []byte, peeridOrConn) {}

package p2ptransport

package udptransport

package manet

## TODO

- [ ] go-multiaddr-net datagram support
- [ ] go-libp2p-transport datagram support
- [ ] go-udp-transport
- [ ] go-peerstream datagram support
- [ ] go-libp2p-swarm datagram support
- [ ] go-multigram
- [ ] go-shs vs. go-dtls vs. go-cryptoauth
- [ ] label-based switch
- [ ] overlay routing interfaces


## PR

This PR adds APIs for message-based communication
