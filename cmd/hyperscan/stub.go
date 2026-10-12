//go:build !hyperscan || !cgo

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Hyperscan requires CGO_ENABLED=1, -tags hyperscan and libhyperscan-dev; use just docker-build and just hyperscan")
	os.Exit(1)
}
