package resolver

import (
	"os"
	"strings"
)

func ResolveSeedNodes() []string {
	seedNodes := os.Getenv("SEED_NODES")

	dnsSeeds := strings.Split(seedNodes, ",")

	return dnsSeeds
}
