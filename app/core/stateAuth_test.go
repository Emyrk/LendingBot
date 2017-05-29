package core_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/app/core"
)

func TestUserAuth(t *testing.T) {
	s := NewStateWithMap()
	err := s.NewUser("testing", "testing")
	if err != nil {
		t.Error(err)
	}

	v, u2, err := s.AuthenticateUser("testing", "testing")
	if err != nil {
		t.Error(err)
	}

	if !v {
		t.Error("User did not validate")
	}

	u, err := s.FetchUser("testing")
	if err != nil {
		t.Error(err)
		if !u2.IsSameAs(u) {
			t.Error("User is not the same")
		}
	}
}
