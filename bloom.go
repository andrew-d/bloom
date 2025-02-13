// Package bloom contains a generic bloom filter type that can be used with any
// comparable value.
package bloom

import (
	"hash/maphash"
	"math"
)

// Filter represents a space-efficient probabilistic data structure
// that tests whether an element is a member of a set.
type Filter[T comparable] struct {
	bits    []uint64
	k       uint           // number of hash functions
	m       uint           // size of bit array
	seeds   []maphash.Seed // k different seeds for k hash functions
	hasher  maphash.Hash
	entries uint
}

// NewBloomFilter creates a new Bloom filter optimized for the expected number
// of items and desired false positive rate.
func NewBloomFilter[T comparable](expectedItems uint, falsePositiveRate float64) *Filter[T] {
	// Calculate optimal size and number of hash functions
	m := optimalBitArraySize(expectedItems, falsePositiveRate)
	k := optimalHashFunctions(expectedItems, m)

	// Generate k different seeds
	seeds := make([]maphash.Seed, k)
	for i := range seeds {
		seeds[i] = maphash.MakeSeed()
	}

	bf := &Filter[T]{
		bits:    make([]uint64, (m+63)/64), // Round up to nearest multiple of 64
		k:       k,
		m:       m,
		seeds:   seeds,
		hasher:  maphash.Hash{},
		entries: 0,
	}
	return bf
}

// Add inserts an item into the Bloom filter.
func (bf *Filter[T]) Add(item T) {
	bf.entries++

	// Use k different hash functions with different seeds
	for i := uint(0); i < bf.k; i++ {
		hash := bf.hashItem(item, bf.seeds[i])
		combinedHash := hash % uint64(bf.m)
		wordIndex := combinedHash / 64
		bitOffset := combinedHash % 64
		bf.bits[wordIndex] |= 1 << bitOffset
	}
}

// Contains tests whether an item might be in the set.
// False positives are possible, but false negatives are not.
func (bf *Filter[T]) Contains(item T) bool {
	// Check all k positions
	for i := uint(0); i < bf.k; i++ {
		hash := bf.hashItem(item, bf.seeds[i])
		combinedHash := hash % uint64(bf.m)
		wordIndex := combinedHash / 64
		bitOffset := combinedHash % 64
		if bf.bits[wordIndex]&(1<<bitOffset) == 0 {
			return false
		}
	}
	return true
}

// hashItem generates a hash value using the provided seed
func (bf *Filter[T]) hashItem(item T, seed maphash.Seed) uint64 {
	bf.hasher.Reset()
	bf.hasher.SetSeed(seed)
	maphash.WriteComparable(&bf.hasher, item)
	return bf.hasher.Sum64()
}

// EstimatedFalsePositiveRate returns the current estimated false positive rate
// based on the number of items added.
func (bf *Filter[T]) EstimatedFalsePositiveRate() float64 {
	if bf.entries == 0 {
		return 0
	}
	p := math.Pow(1-math.Exp(-float64(bf.k)*float64(bf.entries)/float64(bf.m)), float64(bf.k))
	return p
}

// optimalBitArraySize calculates the optimal size of the bit array
// given the expected number of items and desired false positive rate.
func optimalBitArraySize(n uint, p float64) uint {
	return uint(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

// optimalHashFunctions calculates the optimal number of hash functions
// given the expected number of items and bit array size.
func optimalHashFunctions(n, m uint) uint {
	k := uint(math.Ceil(float64(m) / float64(n) * math.Log(2)))
	if k < 1 {
		k = 1
	}
	return k
}
