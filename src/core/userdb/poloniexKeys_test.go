package userdb

import (
	"testing"

	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestExchKeys(t *testing.T) {
	var c [32]byte

	access := "UtyAYBrei4riqry9pkD1iul71A8rBYX5Vc1rNYOCgJH"
	secret := "7l1XzQqnrLM5fVjKDWwkcYJvldvAZz5VwrZ1zwBT1jo"

	pk, err := NewExchangeKeys(access, secret, c)
	if err != nil {
		t.Error(err)
	}

	api, err := pk.DecryptAPIKeyString(c)
	if err != nil {
		t.Error(err)
	}

	sec, err := pk.DecryptAPISecret(c)
	if err != nil {
		t.Error(err)
	}

	if string(sec) != secret {
		t.Errorf("Secret exp %s found %s", secret, sec)
	}

	if string(api) != access {
		t.Errorf("Secret exp %s found %s", access, api)
	}

}
