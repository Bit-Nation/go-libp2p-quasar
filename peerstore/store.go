package peerstore

import (
	"sync"

	net "gx/ipfs/QmYj8wdn5sZEHX2XMDWGBvcXJNdzVbaVpHmXvhHBVZepen/go-libp2p-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func New() *PeerStore {
	return &PeerStore{
		peers: map[peer.ID]net.Stream{},
		lock:  sync.Mutex{},
	}
}

type PeerStore struct {
	peers map[peer.ID]net.Stream
	lock  sync.Mutex
}

func (p *PeerStore) Add(str net.Stream) {
	p.lock.Lock()
	p.peers[str.Conn().RemotePeer()] = str
	p.lock.Unlock()
}

func (p *PeerStore) Remove(str net.Stream) {
	p.lock.Lock()
	delete(p.peers, str.Conn().RemotePeer())
	p.lock.Unlock()
}

func (p *PeerStore) All() map[peer.ID]net.Stream {
	p.lock.Lock()
	peers := p.peers
	p.lock.Unlock()
	return peers
}
