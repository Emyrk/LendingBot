package userdb

import (
	"bytes"
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/common/primitives"
	"github.com/DistributedSolutions/LendingBot/app/core/cryption"
)

type PoloniexKeys struct {
	encryptedAPIKey    []byte
	encryptedAPISecret []byte
}

func NewPoloniexKeys(apiKey string, secret string, cipherKey [32]byte) (*PoloniexKeys, error) {
	pk := new(PoloniexKeys)

	cipherBytes, err := cryption.Encrypt([]byte(apiKey), cipherKey[:])
	if err != nil {
		return nil, err
	}
	pk.encryptedAPIKey = cipherBytes

	cipherBytes, err = cryption.Encrypt([]byte(secret), cipherKey[:])
	if err != nil {
		return nil, err
	}
	pk.encryptedAPISecret = cipherBytes

	return pk, nil
}

func (p *PoloniexKeys) DecryptAPIKeyString(cipherKey [32]byte) (APIKey string, err error) {
	k, e := p.DecryptAPIKey(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *PoloniexKeys) DecryptAPISecretString(cipherKey [32]byte) (APISecret string, err error) {
	k, e := p.DecryptAPISecret(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *PoloniexKeys) DecryptAPIKey(cipherKey [32]byte) (APIKey []byte, err error) {
	return cryption.Decrypt(p.encryptedAPIKey, cipherKey[:])
}

func (p *PoloniexKeys) DecryptAPISecret(cipherKey [32]byte) (APISecret []byte, err error) {
	return cryption.Decrypt(p.encryptedAPISecret, cipherKey[:])
}

func (p *PoloniexKeys) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(p.encryptedAPIKey)
	buf.Write(p.encryptedAPISecret)

	return buf.Next(buf.Len()), nil
}
func (p *PoloniexKeys) UnmarshalBinary(data []byte) (err error) {
	_, err = p.UnmarshalBinaryData(data)
	return err
}

func (p *PoloniexKeys) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[PoloniexKeys] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data
	var b []byte

	b, newData, err = primitives.UnmarshalBinarySliceData(newData)
	if err != nil {
		return data, err
	}
	p.encryptedAPIKey = b

	b, newData, err = primitives.UnmarshalBinarySliceData(newData)
	if err != nil {
		return data, err
	}
	p.encryptedAPISecret = b
	return
}
