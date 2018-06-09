package quasar

import (
	"context"
	"sync"
	"time"

	atbf "github.com/Bit-Nation/go-libp2p-quasar/atbf"
	ps "github.com/Bit-Nation/go-libp2p-quasar/peerstore"
	log "github.com/ipfs/go-log"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
)

var logger = log.Logger("quasar")

// bloom filter size
const B = uint(20000)

const K = 3

const Depth = 3

const ProtocolID = "/quasar/1.0.0"

type Quasar struct {
	dht           *dht.IpfsDHT
	host          host.Host
	peerStore     *ps.PeerStore
	subscriptions map[string]bool
	protocol      *protocol
	filter        *atbf.AttenuatedBloomFilter
	lock          sync.Mutex
}

// create a new instance of quasar
func New(host host.Host) *Quasar {

	peerStore := ps.New()

	filter := atbf.New(Depth, B, K)

	proto := &protocol{
		lock:      sync.Mutex{},
		requests:  map[string]request{},
		filter:    filter,
		peerStore: peerStore,
	}

	// register stream handler
	host.SetStreamHandler(ProtocolID, proto.streamHandler)

	return &Quasar{
		dht:           nil,
		host:          host,
		peerStore:     peerStore,
		protocol:      proto,
		subscriptions: map[string]bool{},
		filter:        filter,
		lock:          sync.Mutex{},
	}
}

// dial to peer on quasar protocol
func (q *Quasar) Dial(ctx context.Context, p peer.ID) error {
	stream, err := q.host.NewStream(ctx, p, ProtocolID)
	if err == nil {
		q.protocol.streamHandler(stream)
		q.peerStore.Add(stream)
	}
	return err
}

// update the filters in the network
// The part that update the network is supposed to have
// a time limit. So that you don't pull filters every time
// you update your filter. However, we will not have that timeout
// since bandwidth isn't a problem for us (ATM)
func (q *Quasar) Commit() error {

	peers := q.peerStore.All()

	// Clear my filter
	if err := q.filter.ClearMyFilter(); err != nil {
		return err
	}

	// Rebuild my filter
	if err := q.filter.Add(0, []byte(q.host.ID().Pretty())); err != nil {
		return err
	}
	for sub, _ := range q.subscriptions {
		if err := q.filter.Add(0, []byte(sub)); err != nil {
			return err
		}
	}

	wg := sync.WaitGroup{}

	// fetch filters from my neighbors
	// and merge them into our filter
	for _, str := range peers {

		wg.Add(1)

		go func(str net.Stream) {
			defer wg.Done()

			// fetch filter from contact
			filter, err := q.protocol.pullFilters(str, time.Second*15)
			if err != nil {
				logger.Error(err)
			}

			// received filter
			recFilter := &atbf.AttenuatedBloomFilter{}
			if err := recFilter.Unmarshal(filter.AttenuateBloomFilter); err != nil {
				logger.Error(err)
				return
			}

			// merge received filter into our filter
			if err := q.protocol.filter.Merge(recFilter); err != nil {
				logger.Error(err)
				return
			}
		}(str)

	}

	wg.Wait()

	// now publish our fresh filter
	wg = sync.WaitGroup{}
	for _, p := range peers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := q.protocol.pushFilters(p, time.Second*15, q.filter)
			if err != nil {
				logger.Error(err)
			}
		}()
	}
	wg.Wait()

	return nil
}

// subscribe to topic
// you have to call "commit" in order to update the network
func (q *Quasar) Subscribe(topic string) {
	// add subscription id
	q.lock.Lock()
	defer q.lock.Unlock()
	q.subscriptions[topic] = false
}

// unsubscribe from topic
// you have to call "commit" in order to update the network
func (q *Quasar) Unsubscribe(id string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	delete(q.subscriptions, id)
}
