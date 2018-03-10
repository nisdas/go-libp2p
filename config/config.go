package config

import (
	"context"
	"crypto/rand"

	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"

	logging "github.com/ipfs/go-log"
	circuit "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	ifconnmgr "github.com/libp2p/go-libp2p-interface-connmgr"
	pnet "github.com/libp2p/go-libp2p-interface-pnet"
	metrics "github.com/libp2p/go-libp2p-metrics"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	filter "github.com/libp2p/go-maddr-filter"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("p2p-config")

// AddrsFactory is a function that takes a set of multiaddrs we're listening on and
// returns the set of multiaddrs we should advertise to the network.
type AddrsFactory = bhost.AddrsFactory

// NATManagerC is a NATManager constructor.
type NATManagerC func(inet.Network) bhost.NATManager

// Config describes a set of settings for a libp2p node
type Config struct {
	Transports         []TptC
	Muxers             []MsMuxC
	SecurityTransports []MsSecC
	ListenAddrs        []ma.Multiaddr
	PeerKey            crypto.PrivKey
	Peerstore          pstore.Peerstore
	Protector          pnet.Protector
	Reporter           metrics.Reporter
	Relay              bool
	RelayOpts          []circuit.RelayOpt
	Insecure           bool
	ConnManager        ifconnmgr.ConnManager
	AddrsFactory       bhost.AddrsFactory
	Filters            *filter.Filters
	NATManager         NATManagerC
}

// NewNode constructs a new libp2p Host based from the Config.
func (cfg *Config) NewNode(ctx context.Context) (host.Host, error) {
	// Check this early. Prevents us from even *starting* without verifying this.
	if pnet.ForcePrivateNetwork && cfg.Protector == nil {
		log.Error("tried to dial with no Private Network Protector but usage" +
			" of Private Networks is forced by the enviroment")
		// Note: This is *also* checked the upgrader itself so it'll be
		// enforced even *if* you don't use the libp2p constructor.
		return nil, pnet.ErrNotInPrivateNetwork
	}

	// If no key was given, generate a random 2048 bit RSA key
	privKey := cfg.PeerKey
	if cfg.PeerKey == nil {
		priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		if err != nil {
			return nil, err
		}
		privKey = priv
	}

	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(privKey.GetPublic())
	if err != nil {
		return nil, err
	}

	// Create a new blank peerstore if none was passed in
	ps := cfg.Peerstore
	if ps == nil {
		ps = pstore.NewPeerstore()
	}
	if !cfg.Insecure {
		ps.AddPrivKey(pid, cfg.PeerKey)
		ps.AddPubKey(pid, cfg.PeerKey.GetPublic())
	}

	// TODO: Make the swarm implementation configurable.
	swrm := swarm.NewSwarm(ctx, pid, ps, cfg.Reporter)
	if cfg.Filters != nil {
		swrm.Filters = cfg.Filters
	}

	// TODO: make host implementation configurable.
	h, err := bhost.NewHost(ctx, swrm, &bhost.HostOpts{
		ConnManager:  cfg.ConnManager,
		AddrsFactory: cfg.AddrsFactory,
		NATManager:   cfg.NATManager,
	})
	if err != nil {
		swrm.Close()
		return nil, err
	}

	upgrader := new(tptu.Upgrader)
	upgrader.Protector = cfg.Protector
	upgrader.Filters = swrm.Filters
	if cfg.Insecure {
		upgrader.Secure = makeInsecureTransport(pid)
	} else {
		upgrader.Secure, err = makeSecurityTransport(h, cfg.SecurityTransports)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	upgrader.Muxer, err = makeMuxer(h, cfg.Muxers)
	if err != nil {
		h.Close()
		return nil, err
	}

	tpts, err := makeTransports(h, upgrader, cfg.Transports)
	if err != nil {
		h.Close()
		return nil, err
	}
	for _, t := range tpts {
		err = swrm.AddTransport(t)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	if cfg.Relay {
		err := circuit.AddRelayTransport(swrm.Context(), h, upgrader, cfg.RelayOpts...)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	// TODO: This method succeeds if listening on one address succeeds. We
	// should probably fail if listening on *any* addr fails.
	if err := h.Network().Listen(cfg.ListenAddrs...); err != nil {
		h.Close()
		return nil, err
	}

	// TODO: Configure routing (it's a pain to setup).
	// TODO: Bootstrapping.

	return h, nil
}

// Option is a libp2p config option that can be given to the libp2p constructor
// (`libp2p.New`).
type Option func(cfg *Config) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}
