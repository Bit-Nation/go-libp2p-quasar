package atbf

import (
	"errors"
	"fmt"
	"sync"
	"encoding/json"
	
	bloom "github.com/willf/bloom"
	"sort"
)

// Create new AttenuatedBloomFilter
func New(depth, m, k uint) *AttenuatedBloomFilter {

	f := &AttenuatedBloomFilter{
		filters: map[uint]*bloom.BloomFilter{},
		lock:    sync.Mutex{},
	}

	start := uint(0)
	for start <= depth {
		f.filters[start] = bloom.New(m, k)
		start++
	}
	return f
}

type AttenuatedBloomFilter struct {
	lock    sync.Mutex
	filters map[uint]*bloom.BloomFilter
}

type attenuatedBloomFilter struct {
	Filters map[uint]*bloom.BloomFilter `json:"filters"`
}

// add to filter
func (b *AttenuatedBloomFilter) Add(depth uint, data []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	filter, exist := b.filters[depth]
	if !exist {
		return errors.New(fmt.Sprintf("invalid depth: %d", depth))
	}
	filter.TestAndAdd(data)
	return nil
}

// check if is present in filter
func (b *AttenuatedBloomFilter) Test(depth uint, data []byte) (bool, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	filter, exist := b.filters[depth]
	if !exist {
		return false, errors.New(fmt.Sprintf("invalid depth: %d", depth))
	}
	exist = filter.Test(data)
	return exist, nil
}

// get filter by depth
func (b *AttenuatedBloomFilter) Get(depth uint) (*bloom.BloomFilter, error) {
	b.lock.Lock()
	filter, exist := b.filters[depth]
	b.lock.Unlock()
	if !exist {
		return nil, errors.New(fmt.Sprintf("invalid depth: %d", depth))
	}
	return filter, nil
}

// merge an remote attenuated bloom filter
func (b *AttenuatedBloomFilter) Merge(remoteFilter *AttenuatedBloomFilter) error {

	b.lock.Lock()
	defer b.lock.Unlock()
	
	// To store the keys in slice in sorted order
	var keys []int
	for k := range b.filters {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	
	// To perform the opertion you want
	for _, k := range keys {
		remFilter, exist := remoteFilter.filters[uint(k)]
		if !exist {
			return errors.New("couldn't find filter tho it should exist")
		}
		
		// 0th filter is always our own one so we don't want to merge them
		// therefore we are shifting one to the right
		// 0th of the remote filter is gonna be our 1th
		myFilter, exist := b.filters[uint(k)+1]
		if !exist {
			break
		}
		myFilter.Merge(remFilter)
		
	}
	
	return nil

}

func (b *AttenuatedBloomFilter) Marshal() ([]byte, error) {
	return json.Marshal(attenuatedBloomFilter{
		Filters: b.filters,
	})
}

func (b *AttenuatedBloomFilter) Unmarshal(data []byte) error  {
	var f attenuatedBloomFilter
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	b.filters = f.Filters
	b.lock = sync.Mutex{}
	return nil
}
