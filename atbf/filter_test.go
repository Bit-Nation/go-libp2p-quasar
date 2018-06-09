package atbf

import (
	"testing"

	require "github.com/stretchr/testify/require"
)

func TestAttenuatedBloomFilter_AddAndTest(t *testing.T) {

	f := New(2, 3, 3)

	// Test that adding a element works
	require.Nil(t, f.Add(0, []byte("hi")))
	exist, err := f.Test(0, []byte("hi"))
	require.Nil(t, err)
	require.True(t, exist)

	// Adding an element to invalid depth shouldn't work since the filter is not present
	require.EqualError(t, f.Add(3, []byte("hi")), "invalid depth: 3")

}

func TestAttenuatedBloomFilter_Merge(t *testing.T) {

	myFilter := New(2, 3, 3)
	remoteFilter := New(2, 3, 3)

	// add a test element
	require.Nil(t, remoteFilter.Add(0, []byte("hi")))

	// merge it
	myFilter.Merge(remoteFilter)

	exist, err := myFilter.Test(1, []byte("hi"))
	require.Nil(t, err)
	require.True(t, exist)

}

func TestJsonMarshalAndUnmarshal(t *testing.T) {

	// create filter
	filter := New(4, 4, 4)

	// add item to filter
	require.Nil(t, filter.Add(3, []byte("hi")))
	exist, err := filter.Test(3, []byte("hi"))
	require.Nil(t, err)
	require.True(t, exist)

	// export the whole thing
	exported, err := filter.Marshal()
	require.Nil(t, err)

	// unmarshal filter
	recoveredFilter := AttenuatedBloomFilter{}
	require.Nil(t, recoveredFilter.Unmarshal(exported))

	// test if filter exist
	exist, err = recoveredFilter.Test(3, []byte("hi"))
	require.Nil(t, err)
	require.True(t, exist)

}

func TestAttenuatedBloomFilter_ClearMyFilter(t *testing.T) {

	// create filter
	filter := New(4, 4, 4)

	// add a to filter
	require.Nil(t, filter.Add(0, []byte("a")))
	exist, err := filter.Test(0, []byte("a"))
	require.Nil(t, err)
	require.True(t, exist)

	// add b to filter
	require.Nil(t, filter.Add(1, []byte("b")))
	exist, err = filter.Test(1, []byte("b"))
	require.Nil(t, err)
	require.True(t, exist)

	// clear only my filter and make sure elements are removed
	filter.ClearMyFilter()
	exist, err = filter.Test(0, []byte("a"))
	require.Nil(t, err)
	require.False(t, exist)

	// check if elements in other filters are still there
	exist, err = filter.Test(1, []byte("b"))
	require.Nil(t, err)
	require.True(t, exist)

}
