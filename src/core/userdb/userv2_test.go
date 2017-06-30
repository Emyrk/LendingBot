package userdb_test

import (
	//	"bytes"
	"crypto/rand"
	"encoding/json"
	"testing"

	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestJsonMarshal(t *testing.T) {
	sec := make([]byte, 32)
	rand.Read(sec)
	var fixed [32]byte
	copy(fixed[:32], sec[:32])

	u, err := NewUser("u@u.com", "u")
	if err != nil {
		t.Error(err)
	}

	u.PoloniexKeys, err = NewExchangeKeys("Random", "WhoCares", fixed)
	if err != nil {
		t.Error(err)
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Error(err)
	}

	u2 := new(User)
	err = json.Unmarshal(data, u2)
	if err != nil {
		t.Error(err)
	}

	apisec, err := u2.PoloniexKeys.DecryptAPISecretString(fixed)
	if err != nil {
		t.Error(err)
	}

	if apisec != "WhoCares" {
		t.Error("BAd sec key")
	}

	apipub, err := u2.PoloniexKeys.DecryptAPIKeyString(fixed)
	if err != nil {
		t.Error(err)
	}

	if apipub != "Random" {
		t.Error("Bad pub key")
	}
}
