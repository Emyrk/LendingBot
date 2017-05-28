package userdb_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/app/core/userdb"
)

func TestUserMarshal(t *testing.T) {
	u, err := NewUser("1", "2")
	if err != nil {
		t.Error(err)
	}

	data, err := u.MarshalBinary()
	if err != nil {
		t.Error(err)
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
