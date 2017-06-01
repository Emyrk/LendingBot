package userdb_test

import (
	"crypto"
	"fmt"
	"os"
	"testing"

	"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestUserStat(t *testing.T) {
	u, err := new(UserStatistic)
	if err != nil {
		t.Error(err)
	}

	u.PoloniexEnabled = true

	data, err := u.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if !u.PoloniexKeys.APIKeyEmpty() {
		t.Error("Should be empty")
	}

	if !u.PoloniexKeys.SecretKeyEmpty() {
		t.Error("Should be empty")
	}

	u2 := NewBlankUser()
	nd, err := u2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	if len(nd) != 0 {
		t.Errorf("%d bytes left", len(nd))
	}

	if !u.IsSameAs(u2) {
		t.Error("Should be same")
	}
}
