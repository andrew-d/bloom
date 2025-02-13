// Package bloom contains a generic bloom filter type that can be used with any
// comparable value.
package bloom

import (
	"hash/maphash"
	"math"
)

// Filter represents a space-efficient probabilistic data structure that tests
// whether an element is a member of a set. Once a Filter has been created,
// adding a new item to the set does not require any additional memory
// allocation.
//
// None of the methods on this type are safe for concurrent use.
type Filter[T comparable] struct {
	bits    []uint64
	m       uint           // size of bit array
	seeds   []maphash.Seed // k different seeds for k hash functions
	hasher  maphash.Hash
	entries uint
}

// NewBloomFilter creates a new Bloom filter optimized for the expected number
// of items and desired false positive rate.
func NewBloomFilter[T comparable](expectedItems uint, falsePositiveRate float64) *Filter[T] {
	// Calculate optimal size and number of hash functions
	m, k := bloomParams(expectedItems, falsePositiveRate)

	// Generate k different seeds
	seeds := make([]maphash.Seed, k)
	for i := range seeds {
		seeds[i] = maphash.MakeSeed()
	}

	bf := &Filter[T]{
		bits:    make([]uint64, (m+63)/64), // Round up to nearest multiple of 64
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

	// Set a bit for each of our hash functions.
	for _, seed := range bf.seeds {
		hash := bf.hashItem(item, seed)
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
	for _, seed := range bf.seeds {
		hash := bf.hashItem(item, seed)
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

	// From Wikipedia: https://en.wikipedia.org/wiki/Bloom_filter#Probability_of_false_positives
	//
	//  1. The probability of any single bit is not set to 1 by any of our
	//     k hash functions is:
	//     (1 - 1/m)^k ≈ e^(-k/m)
	//
	//  2. If we have inserted n items, the probability of a certain bit is
	//     still 0 is:
	//     e^(-kn/m)
	//
	//  3. The probability that the bit is 1 is:
	//     1 - e^(-kn/m)
	//
	//  4. Now, for a false positive to occur, we need all k bits to be set
	//     for a given element that's not in the set. The probability that
	//     this occurs for all k bits is:
	//     (1 - e^(-kn/m))^k

	k := float64(len(bf.seeds))
	n := float64(bf.entries)
	m := float64(bf.m)

	probBitIsZero := math.Exp(-k * n / m)
	probBitIsOne := 1 - probBitIsZero
	return math.Pow(probBitIsOne, k)
}

func bloomParams(expectedItems uint, falsePositiveRate float64) (bitsNeeded uint, numHashFunctions uint) {
	// Use the standard naming from Wikipedia to make the equations easier to follow
	n := float64(expectedItems)
	p := falsePositiveRate

	// Calculate the number of bits we need in our bit array
	m := -n * math.Log(p) / math.Pow(math.Log(2), 2)
	bitsNeeded = uint(math.Ceil(m))

	// Calculate the number of hash functions we need
	k := m / n * math.Log(2)
	numHashFunctions = uint(math.Ceil(k))

	// Clamp to at least 1 hash function
	numHashFunctions = max(numHashFunctions, 1)
	return
}
