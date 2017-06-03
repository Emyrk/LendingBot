package core_test

import (
	"fmt"
	"testing"

	. "github.com/Emyrk/LendingBot/src/core"
)

var _ = fmt.Println

func TestStateDBVerify(t *testing.T) {
	s := NewStateWithMap()
	err := s.VerifyState()
	if err != nil {
		t.Error(err)
	}

	s.CipherKey[6] = 0xFF
	err = s.VerifyState()
	if err == nil {
		t.Error("Should error")
	}

}

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

	err = s.VerifyEmail("testing", u.VerifyString)
	if err != nil {
		t.Error(err)
	}
}

func TestUserKeys(t *testing.T) {
	s := NewStateWithMap()
	err := s.NewUser("testing", "testing")
	if err != nil {
		t.Error(err)
	}

	u, err := s.FetchUser("testing")
	if err != nil {
		t.Error(err)
	}

	if !u.PoloniexKeys.APIKeyEmpty() {
		t.Error("Should be empty")
	}
	//fmt.Printf("%t\n", u.PoloniexKeys.APIKeyEmpty())
	s.SetUserKeys("testing", "accesskey", "secretkey")

	u, err = s.FetchUser("testing")
	if err != nil {
		t.Error(err)
	}

	if u.PoloniexKeys.APIKeyEmpty() {
		t.Error("Should not be empty")
	}

	//fmt.Printf("%t\n", u.PoloniexKeys.APIKeyEmpty())
	// fmt.Println(u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(s.CipherKey)))
	if m, _ := u.PoloniexKeys.DecryptAPIKeyString(u.GetCipherKey(s.CipherKey)); m != "accesskey" {
		t.Error("bad decrypt")
	}
}
