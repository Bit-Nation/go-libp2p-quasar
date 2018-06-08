package atbf

import (
	"errors"
	"fmt"
	"sync"

	bloom "github.com/willf/bloom"
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
func (b *AttenuatedBloomFilter) Merge(remoteFilter *AttenuatedBloomFilter) {

	b.lock.Lock()
	defer b.lock.Unlock()

	for index, remFilter := range remoteFilter.filters {
		// 0th filter is always our own one so we don't want to merge them
		// therefore we are shifting one to the right
		// 0th of the remote filter is gonna be our 1th
		myFilter, exist := b.filters[index+1]
		if !exist {
			break
		}
		myFilter.Merge(remFilter)
	}

}
