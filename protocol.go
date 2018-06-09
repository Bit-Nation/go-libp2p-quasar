package quasar

import (
	"sync"
	
	atbf "github.com/Bit-Nation/go-libp2p-quasar/atbf"
	ps "github.com/Bit-Nation/go-libp2p-quasar/peerstore"
)

type reads struct{}

type protocol struct {
	lock      sync.Mutex
	peerStore *ps.PeerStore
	requests  map[string]request
	filter    *atbf.AttenuatedBloomFilter
}

func NewProto() *protocol {

	return &protocol{
		requests:  map[string]request{},
		lock:      sync.Mutex{},
		peerStore: ps.New(),
	}

}
