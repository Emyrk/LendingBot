package core_test

import (
	"fmt"
	"testing"

	. "github.com/Emyrk/LendingBot/src/core"
)

var _ = fmt.Println

func TestValidEmail(t *testing.T) {
	emails := make([]struct {
		Email string
		Bad   bool
	}, 20)
	emails[0].Email = "email@example.com"
	emails[1].Email = "firstname.lastname@example.com"
	emails[2].Email = "email@subdomain.example.com"
	emails[3].Email = "steven@gmail.com"
	emails[4].Email = "steven@facbook.com"
	emails[5].Email = "a@a.com"
	emails[6].Email = "bob@mail.com"
	emails[7].Email = "chris@hello.com"
	emails[8].Email = "peter@google.com"
	emails[9].Email = "lol@a.net"
	emails[10].Email = "a@rit.edu"
	emails[11].Email = "billy@"
	emails[11].Bad = true
	emails[12].Email = "plainaddress"
	emails[12].Bad = true
	emails[13].Email = "@example.com"
	emails[13].Bad = true
	emails[14].Email = "email@example.com (Joe Smith)"
	emails[14].Bad = true
	emails[15].Email = "emailexample.web.web.web"
	emails[15].Bad = true
	emails[16].Email = "email111.222.333.44444"
	emails[16].Bad = true
	emails[17].Email = "Abc..123@@example"
	emails[17].Bad = true
	emails[18].Email = `“(),:;<>[\]@example.com`
	emails[18].Bad = true
	emails[19].Email = `this\ is"really"not\allowed@example.com`
	emails[19].Bad = true

	for _, v := range emails {
		err := ValidateEmail(v.Email)
		if v.Bad {
			if err == nil {
				t.Errorf("Email '%s' came back valid, it should not be", v.Email)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
		}
	}
}

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
