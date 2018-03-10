package config

import (
	"fmt"
	"reflect"

	host "github.com/libp2p/go-libp2p-host"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	mux "github.com/libp2p/go-stream-muxer"
	mplex "github.com/whyrusleeping/go-smux-multiplex"
	msmux "github.com/whyrusleeping/go-smux-multistream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
)

// MuxC is a stream multiplex transport constructor
type MuxC func(h host.Host) (mux.Transport, error)

// MsMuxC is a tuple containing a multiplex transport constructor and a protocol
// ID.
type MsMuxC struct {
	MuxC
	ID string
}

var muxArgTypes = map[reflect.Type]constructor{
	hostType:    func(h host.Host, _ *tptu.Upgrader) interface{} { return h },
	networkType: func(h host.Host, _ *tptu.Upgrader) interface{} { return h.Network() },
	peerIDType:  func(h host.Host, _ *tptu.Upgrader) interface{} { return h.ID() },
	pstoreType:  func(h host.Host, _ *tptu.Upgrader) interface{} { return h.Peerstore() },
}

// MuxerConstructor creates a multiplex constructor from the passed parameter
// using reflection.
func MuxerConstructor(m interface{}) (MuxC, error) {
	// Already constructed?
	if t, ok := m.(mux.Transport); ok {
		return func(_ host.Host) (mux.Transport, error) {
			return t, nil
		}, nil
	}

	fn, err := makeConstructor(m, muxType, muxArgTypes)
	if err != nil {
		return nil, err
	}
	return func(h host.Host) (mux.Transport, error) {
		t, err := fn(h, nil)
		if err != nil {
			return nil, err
		}
		return t.(mux.Transport), nil
	}, nil
}

func makeMuxer(h host.Host, tpts []MsMuxC) (mux.Transport, error) {
	muxMuxer := msmux.NewBlankTransport()
	if len(tpts) == 0 {
		transportSet := make(map[string]struct{}, len(tpts))
		for _, tptC := range tpts {
			if _, ok := transportSet[tptC.ID]; ok {
				return nil, fmt.Errorf("duplicate muxer transport: %s", tptC.ID)
			}
		}
		for _, tptC := range tpts {
			tpt, err := tptC.MuxC(h)
			if err != nil {
				return nil, err
			}
			muxMuxer.AddTransport(tptC.ID, tpt)
		}
	} else {
		muxMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
		muxMuxer.AddTransport("/mplex/6.3.0", mplex.DefaultTransport)
	}
	return muxMuxer, nil
}
