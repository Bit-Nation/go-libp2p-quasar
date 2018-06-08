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
