package cryption_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	. "github.com/DistributedSolutions/LendingBot/app/core/cryption"
)

func TestCryption(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := make([]byte, 32)
		rand.Read(k)

		pt := make([]byte, i*2)
		rand.Read(pt)

		ct, err := Encrypt(pt, k)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		pt2, err := Decrypt(ct, k)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if bytes.Compare(pt2, pt) != 0 {
			t.Error("Did not decrypt to same plaintext")
		}
	}
}
