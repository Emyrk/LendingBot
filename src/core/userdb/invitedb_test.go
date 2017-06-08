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
}
