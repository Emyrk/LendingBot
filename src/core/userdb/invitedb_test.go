package userdb_test

import (
	"testing"
	"time"

	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestInviteCode(t *testing.T) {
	id := NewInviteMapDB()
	err := id.CreateInviteCode("Code1", 1, time.Now().Add(1000000*time.Hour))
	if err != nil {
		t.Error(err)
	}

	good, err := id.ClaimInviteCode("user1", "Code1")
	if err != nil {
		t.Error(err)
	}

	if !good {
		t.Errorf("Should have worked")
	}

	good, err = id.ClaimInviteCode("user2", "Code1")
	if err == nil || good {
		t.Error("Should error")
	}

	v, err := id.ListAll()
	t.Log(v, err)
}

func TestInviteEntry(t *testing.T) {
	ie := NewInviteCode("Hello", 2, time.Now())
	data, err := ie.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	ie2 := new(InviteEntry)
	nd, err := ie2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	if len(nd) > 0 {
		t.Error("Should be 0")
	}

	if ie2.RawCode != ie.RawCode {
		t.Error("Raw code does not match")
	}
	t.Log(ie2, ie)
}
