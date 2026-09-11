// Package names generates short, human-readable share identifiers.
package names

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var words = []string{
	"bluebird", "cobalt", "maple", "orbit", "juniper", "willow", "ember",
	"harbor", "meadow", "cedar", "canyon", "lagoon", "ridge", "summit",
	"driftwood", "granite", "hazel", "indigo", "jasper", "kestrel",
	"lantern", "mistral", "nectar", "opal", "pebble", "quartz", "raven",
	"sable", "thistle", "umber", "violet", "walnut", "yarrow", "zephyr",
	"amber", "birch", "clover", "delta", "ferny", "grove", "heron",
	"ivory", "jade", "koala", "lotus", "marlin", "nimbus", "otter",
	"prairie", "quill", "reef", "spruce", "tundra", "urchin", "vale",
	"wren", "yonder", "zenith", "aspen", "basin", "coral", "dune",
}

// Generate returns a random word not already present in taken.
func Generate(taken map[string]bool) (string, error) {
	for attempt := 0; attempt < 200; attempt++ {
		w, err := pick()
		if err != nil {
			return "", err
		}
		if attempt >= 40 {
			// Word pool exhausted; disambiguate with a short numeric suffix.
			n, err := rand.Int(rand.Reader, big.NewInt(90))
			if err != nil {
				return "", err
			}
			w = fmt.Sprintf("%s%d", w, n.Int64()+10)
		}
		if !taken[w] {
			return w, nil
		}
	}
	return "", fmt.Errorf("names: could not find a free name")
}

func pick() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		return "", err
	}
	return words[n.Int64()], nil
}
