package userdb_test

import (
	"crypto"
	"os"
	"testing"

	"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/app/core/userdb"
)

func TestUserMarshal(t *testing.T) {
	u, err := NewUser("1", "2")
	if err != nil {
		t.Error(err)
	}

	u.PoloniexEnabled = true

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

func Test2FA(t *testing.T) {
	otp, err := twofactor.NewTOTP("info@sec51.com", "Sec51", crypto.SHA256, 8)
	if err != nil {
		t.Error(err)
	}

	b, err := otp.ToBytes()
	if err != nil {
		t.Error(err)
	}

	_, err = twofactor.TOTPFromBytes(b, "Sec51")
	if err != nil {
		t.Error(err)
	}
}

func TestMoreserMarshal(t *testing.T) {
	u, err := NewUser("1asASDASDdsad", "298dfsdfjkhDFdasfsfDSFFFFf")
	if err != nil {
		t.Error(err)
	}

	_, err = u.Create2FA("test")
	if err != nil {
		t.Error(err)
	}

	u.Enabled2FA = true

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

func TestCleanup(t *testing.T) {
	os.RemoveAll("keys")

}
