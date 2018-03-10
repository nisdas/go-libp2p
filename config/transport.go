package config

import (
	"reflect"

	host "github.com/libp2p/go-libp2p-host"
	transport "github.com/libp2p/go-libp2p-transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
)

// TptC is the type for libp2p transport constructors. You probably won't ever
// implement this function interface directly. Instead, pass your transport
// constructor to TransportConstructor.
type TptC func(h host.Host, u *tptu.Upgrader) (transport.Transport, error)

var transportArgTypes = map[reflect.Type]constructor{
	upgraderType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u },
	hostType:      func(h host.Host, u *tptu.Upgrader) interface{} { return h },
	networkType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Network() },
	muxType:       func(h host.Host, u *tptu.Upgrader) interface{} { return u.Muxer },
	securityType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u.Secure },
	protectorType: func(h host.Host, u *tptu.Upgrader) interface{} { return u.Protector },
	filtersType:   func(h host.Host, u *tptu.Upgrader) interface{} { return u.Filters },
	peerIDType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.ID() },
	privKeyType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PrivKey(h.ID()) },
	pubKeyType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PubKey(h.ID()) },
	pstoreType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore() },
}

// TransportConstructor uses reflection to turn a function that constructs a
// transport into a TptC.
//
// You can pass either a constructed transport (something that implements
// `transport.Transport`) or a function that takes any of:
//
// * The local peer ID.
// * A transport connection upgrader.
// * A private key.
// * A public key.
// * A Host.
// * A Network.
// * A Peerstore.
// * An address filter.
// * A security transport.
// * A stream multiplexer transport.
// * A private network protector.
//
// And returns a type implementing transport.Transport and, optionally, an error
// (as the second argument).
func TransportConstructor(tpt interface{}) (TptC, error) {
	// Already constructed?
	if t, ok := tpt.(transport.Transport); ok {
		return func(_ host.Host, _ *tptu.Upgrader) (transport.Transport, error) {
			return t, nil
		}, nil
	}
	fn, err := makeConstructor(tpt, transportType, transportArgTypes)
	if err != nil {
		return nil, err
	}
	return func(h host.Host, u *tptu.Upgrader) (transport.Transport, error) {
		t, err := fn(h, u)
		if err != nil {
			return nil, err
		}
		return t.(transport.Transport), nil
	}, nil
}

func makeTransports(h host.Host, u *tptu.Upgrader, tpts []TptC) ([]transport.Transport, error) {
	if len(tpts) > 0 {
		transports := make([]transport.Transport, len(tpts))
		for i, tC := range tpts {
			t, err := tC(h, u)
			if err != nil {
				return nil, err
			}
			transports[i] = t
		}
		return transports, nil
	}
	return []transport.Transport{tcp.NewTCPTransport(u), ws.New(u)}, nil
}
