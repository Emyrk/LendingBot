package userdb

import (
	"bytes"
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/core/common/primitives"
	"github.com/DistributedSolutions/LendingBot/app/core/cryption"
)

type PoloniexKeys struct {
	EncryptedAPIKey    []byte
	EncryptedAPISecret []byte
}

func NewPoloniexKeys(apiKey string, secret string, cipherKey []byte) (*PoloniexKeys, error) {
	pk := new(PoloniexKeys)

	cipherBytes, err := cryption.Encrypt([]byte(apiKey), cipherKey)
	if err != nil {
		return nil, err
	}
	pk.EncryptedAPIKey = cipherBytes

	cipherBytes, err = cryption.Encrypt([]byte(secret), cipherKey)
	if err != nil {
		return nil, err
	}
	pk.EncryptedAPISecret = cipherBytes

	return pk, nil
}

func (p *PoloniexKeys) DecryptAPIKeyString(cipherKey []byte) (APIKey string, err error) {
	k, e := p.DecryptAPIKey(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *PoloniexKeys) DecryptAPISecretString(cipherKey []byte) (APISecret string, err error) {
	k, e := p.DecryptAPISecret(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *PoloniexKeys) DecryptAPIKey(cipherKey []byte) (APIKey []byte, err error) {
	return cryption.Decrypt(p.EncryptedAPIKey, cipherKey)
}

func (p *PoloniexKeys) DecryptAPISecret(cipherKey []byte) (APISecret []byte, err error) {
	return cryption.Decrypt(p.EncryptedAPISecret, cipherKey)
}

func (p *PoloniexKeys) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(p.EncryptedAPIKey)
	buf.Write(p.EncryptedAPISecret)

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
	p.EncryptedAPIKey = b

	b, newData, err = primitives.UnmarshalBinarySliceData(newData)
	if err != nil {
		return data, err
	}
	p.EncryptedAPISecret = b
	return
}
