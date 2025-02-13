package bloom

import (
	"fmt"
	"strings"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter[string](1000, 0.01)

	// Add some items
	bf.Add("apple")
	bf.Add("banana")
	bf.Add("orange")

	// Test presence
	if !bf.Contains("apple") {
		t.Error("'apple' should be in the filter")
	}
	if !bf.Contains("banana") {
		t.Error("'banana' should be in the filter")
	}
	if bf.Contains("grape") {
		t.Error("'grape' should not be in the filter")
	}
}

func TestBloomFilter_EstimatedFalsePositiveRate(t *testing.T) {
	expectedItems := uint(1000)
	targetFPR := 0.01
	bf := NewBloomFilter[int](expectedItems, targetFPR)

	// The initial false positive rate should be 0, with no items added.
	if fpr := bf.EstimatedFalsePositiveRate(); fpr != 0 {
		t.Errorf("got %v, want 0", fpr)
	}

	// Add expectedItems number of items
	for i := 0; i < int(expectedItems); i++ {
		bf.Add(i)
	}

	// Check that the estimated FPR is close to target
	actualFPR := bf.EstimatedFalsePositiveRate()
	if actualFPR > targetFPR*2 { // Allow some margin
		t.Errorf("false positive rate too high: got %v, want < %v", actualFPR, targetFPR*2)
	}
}

func BenchmarkBloomFilterAdd(b *testing.B) {
	// Test cases with different string lengths
	lengths := []int{10, 100, 1000, 10000}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("length_%d", length), func(b *testing.B) {
			// Create a string of the desired length
			s := strings.Repeat("a", length)

			// Create a new Bloom filter for each test
			bf := NewBloomFilter[string](1000, 0.01)

			b.SetBytes(int64(length))
			b.ReportAllocs()
			for b.Loop() {
				bf.Add(s)
			}
		})
	}
}
