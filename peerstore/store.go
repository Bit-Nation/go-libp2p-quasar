package peerstore

import (
	"sync"

	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
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
