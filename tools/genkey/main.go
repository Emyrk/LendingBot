package main

import (
	"crypto/rand"
	"fmt"
)

func main() {
	key := make([]byte, 32)
	rand.Read(key)

	var fixed [32]byte
	copy(fixed[:32], key[:32])

	fmt.Printf("Private Key: %x\n", fixed[:])
}
