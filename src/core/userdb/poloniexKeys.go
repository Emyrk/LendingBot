package userdb

import (
	"bytes"
	"fmt"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
	"github.com/Emyrk/LendingBot/src/core/cryption"
)

type ExchangeKeys struct {
	EncryptedAPIKey    []byte `json:"EncryptedAPIKey"`
	EncryptedAPISecret []byte `json:"EncryptedAPISecret"`
}

func (a *ExchangeKeys) SetEmptyIfBlank() {
	if a.EncryptedAPIKey == nil {
		a.EncryptedAPIKey = []byte{0x00}
	}
	if a.EncryptedAPISecret == nil {
		a.EncryptedAPISecret = []byte{0x00}
	}
}

func (a *ExchangeKeys) IsSameAs(b *ExchangeKeys) bool {
	if bytes.Compare(a.EncryptedAPIKey, b.EncryptedAPIKey) != 0 {
		return false
	}

	if bytes.Compare(a.EncryptedAPISecret, b.EncryptedAPISecret) != 0 {
		return false
	}

	return true
}

func NewBlankExchangeKeys() *ExchangeKeys {
	return &ExchangeKeys{EncryptedAPIKey: []byte{0x00}, EncryptedAPISecret: []byte{0x00}}
}

func NewExchangeKeys(apiKey string, secret string, cipherKey [32]byte) (*ExchangeKeys, error) {
	pk := new(ExchangeKeys)

	cipherBytes, err := cryption.Encrypt([]byte(apiKey), cipherKey[:])
	if err != nil {
		return nil, err
	}
	pk.EncryptedAPIKey = cipherBytes

	cipherBytes, err = cryption.Encrypt([]byte(secret), cipherKey[:])
	if err != nil {
		return nil, err
	}
	pk.EncryptedAPISecret = cipherBytes

	return pk, nil
}

func (p *ExchangeKeys) APIKeyEmpty() bool {
	return bytes.Compare(p.EncryptedAPIKey, []byte{0x00}) == 0
}

func (p *ExchangeKeys) SecretKeyEmpty() bool {
	return bytes.Compare(p.EncryptedAPISecret, []byte{0x00}) == 0
}

func (p *ExchangeKeys) DecryptAPIKeyString(cipherKey [32]byte) (APIKey string, err error) {
	k, e := p.DecryptAPIKey(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *ExchangeKeys) DecryptAPISecretString(cipherKey [32]byte) (APISecret string, err error) {
	k, e := p.DecryptAPISecret(cipherKey)
	if e != nil {
		return "", e
	}
	return string(k), nil
}

func (p *ExchangeKeys) DecryptAPIKey(cipherKey [32]byte) (APIKey []byte, err error) {
	return cryption.Decrypt(p.EncryptedAPIKey, cipherKey[:])
}

func (p *ExchangeKeys) DecryptAPISecret(cipherKey [32]byte) (APISecret []byte, err error) {
	return cryption.Decrypt(p.EncryptedAPISecret, cipherKey[:])
}

func (p *ExchangeKeys) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	b := primitives.MarshalBinarySlice(p.EncryptedAPIKey)
	buf.Write(b)

	b = primitives.MarshalBinarySlice(p.EncryptedAPISecret)
	buf.Write(b)

	return buf.Next(buf.Len()), nil
}
func (p *ExchangeKeys) UnmarshalBinary(data []byte) (err error) {
	_, err = p.UnmarshalBinaryData(data)
	return err
}

func (p *ExchangeKeys) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[ExchangeKeys] A panic has occurred while unmarshaling: %s", r)
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
