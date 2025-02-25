package analyzers

import (
	"math"
)

// EntropyAnalyzer checks if the entropy of the given content is at least the minimum entropy.
// Returns true if the content passes the check, false otherwise.
func EntropyAnalyzer(content string, config map[string]interface{}) bool {
	minEntropy := 3.0 // default
	if val, ok := config["min"].(float64); ok {
		minEntropy = val
	}

	entropy := calculateShannonEntropy(content)
	return entropy >= minEntropy
}

func calculateShannonEntropy(s string) float64 {
	var entropy float64
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	l := float64(len(s))
	for _, cnt := range counts {
		f := float64(cnt) / l
		entropy -= f * math.Log2(f)
	}
	return entropy
}
