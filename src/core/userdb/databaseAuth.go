package userdb

import (
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Println

func (ud *UserDatabase) GetVerifyString(username string) (string, error) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return "", err
	}
	return u.VerifyString, nil
}

func (ud *UserDatabase) PutVerifystring(usernameHash [32]byte, verifystring string) error {
	key, err := hex.DecodeString(verifystring)
	if err != nil {
		return err
	}

	return ud.db.Put(VerifyBucket, key, usernameHash[:])
}

func (ud *UserDatabase) VerifyEmail(email, verifyString string) error {
	key, err := hex.DecodeString(verifyString)
	if err != nil {
		return err
	}

	uh, err := ud.db.Get(VerifyBucket, key)
	if err != nil {
		return err
	}

	u := NewBlankUser()
	f, err := ud.get(UsersBucket, uh[:], u)
	if err != nil {
		return err
	}

	if !f {
		return fmt.Errorf("User for that string not found")
	}

	if u.VerifyString == verifyString && u.Username == email {
		u.Verified = true
		return ud.PutUser(u)
	}
	return fmt.Errorf("verify string is incorrect")
}

func (ud *UserDatabase) AuthenticateUser(username string, password string, token string) (bool, *User, error) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return false, nil, err
	}

	if !u.AuthenticatePassword(password) {
		return false, nil, nil
	}

	// Passed Password Auth
	if u.Enabled2FA {
		err = ud.validate2FA(u, token)
		if err != nil {
			return false, u, err
		}
	}

	return true, u, nil
}

func (ud *UserDatabase) Add2FA(username string, password string) (qr []byte, err error) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return nil, err
	}

	if u.Enabled2FA {
		return nil, fmt.Errorf("2FA is already enabled. Disable to generate a new barcode")
	}

	if !u.AuthenticatePassword(password) {
		// Only warn if they fail X times
		log.Warnf("%s failed to authenticate", username)
		return nil, fmt.Errorf("Invalid password")
	}

	qr, err = u.Create2FA("HodlZone")
	if err != nil {
		return nil, err
	}

	err = ud.PutUser(u)
	if err != nil {
		return nil, err
	}

	return qr, nil
}

func (ud *UserDatabase) Enable2FA(username string, password string, token string, enabled bool) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	if !u.AuthenticatePassword(password) {
		return fmt.Errorf("Invalid password or 2FA")
	}

	valid := ud.validate2FA(u, token)
	if valid != nil {
		return err
	}

	u.Enabled2FA = enabled
	return ud.PutUser(u)
}

func (ud *UserDatabase) validate2FA(u *User, token string) error {
	return u.Validate2FA(token)
}

func (ud *UserDatabase) UpdateJWTTime(username string, t time.Time) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.JWTTime = t

	return ud.PutUser(u)
}

func (ud *UserDatabase) UpdateJWTOTP(username string, b [32]byte) error {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		return err
	}

	u.ClearJWTOTP()

	err = u.SetJWTOTP(b)
	if err != nil {
		return err
	}

	return ud.PutUser(u)
}

func (ud *UserDatabase) GetJWTOTP(username string) ([32]byte, bool) {
	u, err := ud.FetchUserIfFound(username)
	if err != nil {
		var ret [32]byte
		return ret, false
	}

	return u.GetJWTOTP()
}
