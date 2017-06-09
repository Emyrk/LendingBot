package userdb_test

import (
	"crypto"
	"os"
	"testing"

	"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestUserMarshal(t *testing.T) {
	u, err := NewUser("1", "2")
	if err != nil {
		t.Error(err)
	}

	u.PoloniexEnabled.Enable(true)
	u.LendingStrategy = 10

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

func TestUserWithKeys(t *testing.T) {
	u, err := NewUser("1", "2")
	if err != nil {
		t.Error(err)
	}

	accessKey := "abceaskljfhdfjklfkjsdhfklsdhf"
	secret := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	u.PoloniexEnabled.Enable(true)
	u.LendingStrategy = 10

	var key [32]byte
	u.PoloniexKeys, err = NewPoloniexKeys(accessKey, secret, key)
	if err != nil {
		t.Error(err)
	}

	data, err := u.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if u.PoloniexKeys.APIKeyEmpty() {
		t.Error("Should not be empty")
	}

	if u.PoloniexKeys.SecretKeyEmpty() {
		t.Error("Should not be empty")
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

	v, err := u.PoloniexKeys.DecryptAPIKeyString(key)
	if err != nil {
		t.Error(err)
	}
	if accessKey != v {
		t.Errorf("Got back %s as key, exp %s", v, accessKey)
	}

	v, err = u.PoloniexKeys.DecryptAPISecretString(key)
	if err != nil {
		t.Error(err)
	}
	if secret != v {
		t.Errorf("Got back %s as key, exp %s", v, secret)
	}
}

func TestPE(t *testing.T) {
	pe := new(PoloniexEnabledStruct)
	pe.Enable(PoloniexEnabledStruct{
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
	})

	d := pe.Bytes()
	pe2 := new(PoloniexEnabledStruct)
	nd, err := pe2.UnmarshalBinaryData(d)
	if err != nil {
		t.Error(err)
	}
	if len(nd) > 0 {
		t.Error("Should be 0")
	}
}

func TestMinLend(t *testing.T) {
	m := new(MiniumumLendStruct)
	m.BTC = 0.1

	d, err := m.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	m2 := new(MiniumumLendStruct)
	nd, err := m2.UnmarshalBinaryData(d)
	if err != nil {
		t.Error(err)
	}
	if len(nd) > 0 {
		t.Error("Should be 0")
	}

	if m2.BTC != 0.1 {
		t.Error("Should be 0.1")
	}

}
