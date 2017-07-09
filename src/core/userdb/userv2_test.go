package userdb_test

import (
	//	"bytes"
	"crypto/rand"
	"encoding/hex"
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

func TestUserV2Unmarshal(t *testing.T) {
	data := `73616d2e636f68656e393440676d61696c2e636f6d00ccf035c0d27b75e0caf16dd1d5ded49565343e7a4ee761ce1a4d4c893b6ac05ff5d76498fa010000000ed0cf4ca116e7e2990000010000000ed0cf4ca116e7e3200000000003e5302e31333030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000302e30303030303000000000000100486f646c5a6f6e65000000009d9d0000000000000030f63eea22ec72311abb73c2119f19de73e219a16be70c8f540b8f3799a24e18a37e5a583c84c6c88d0b36c56d4778a3ec4dd587b4d40467aeec9e29710b77e5cc042851184250e14fbbe5bb64cb38a2f40bb7d8ba0c72e7704423dc7b8509e3414238ee57901f410accef8cbb270b08214a5a4311c115a21cd6d236040567818836897ff722b39cb38c9014bd7be23bc8b3d6402f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000137376166306665393762393235613666323063636163343132333532623162336433666633353134623439383231663630363865353139663236343866663932000000000000000000000000000000003fefcd539107e6ea23c82fc28335ca838cc0665b0e5e3698d8b40f5dfd7591821a8e7bd5cd18b3c0f07a255230a262c9af7c9cfa0da6122d02421cc3b9731aae0000009c462183a2ff9054d61f68ce71163f6e4e253043c068a8bfbe832959a3ad49e53d776de48627cba7bdb0831a93ef06d43f9202e03078dbff67f651a70dd3a4906d87bfd010b8f52445f21ec28bf89994987742537fbe400cfef6993be83f8ae7ba107932d883e9a1d45d780f0d705a862ed373b8417cdc6a94a13fd55a6c262df0d806e0e679f70bfec47b7a561325a8ac7b765e206d79bf2035b71804`
	b, err := hex.DecodeString(data)
	if err != nil {
		t.Error(err)
	}

	u := NewBlankUser()
	err = u.SafeUnmarshal(b)
	if err != nil {
		t.Error(err)
	}

	t.Log(u)

}
