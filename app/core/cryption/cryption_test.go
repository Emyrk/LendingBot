package cryption_test

import (
	"bytes"
	"crypto/rand"
	"time"

	"github.com/DistributedSolutions/LendingBot/app/core"
	. "github.com/DistributedSolutions/LendingBot/app/core/cryption"
	"testing"
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

func TestJWT(t *testing.T) {
	start := core.NewState()

	email := "bla@gmail.com"
	tokenString, err := NewJWT(email, start.JWTSecret, JWT_EXPIRY_TIME)
	if err != nil {
		t.Errorf("Error creating jwt: %s\n", err.Error())
	}

	jwtEmail, err := VerifyJWT(tokenString, start.JWTSecret)
	if err != nil {
		t.Errorf("Error verifying jwt: %s\n", err.Error())
	}
	if jwtEmail != email {
		t.Errorf("Error emails do not match: %s, %s\n", jwtEmail, email)
	}

	tokenString, _ = NewJWT(email, start.JWTSecret, JWT_EXPIRY_TIME_TEST_FAIL)

	time.Sleep(3 * time.Second)

	jwtEmail, err = VerifyJWT(tokenString, start.JWTSecret)
	if err == nil {
		t.Error("Error should have produced an error")
	}
	if jwtEmail != "" {
		t.Errorf("Error email should be empty: %s\n", jwtEmail)
	}
}
