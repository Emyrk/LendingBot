package userdb

import (
	"bytes"
	"fmt"

	"github.com/DistributedSolutions/LendingBot/app/common/primitives"
)

type PoloniexKeys struct {
	APIKey    string
	APISecret string
}

func NewPoloniexKeys(key string, secret string) *PoloniexKeys {
	return &PoloniexKeys{APIKey: key, APISecret: secret}
}

func (p *PoloniexKeys) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// TODO: Set a real max length
	b, err := primitives.MarshalStringToBytes(p.APIKey, 1000)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	// TODO: Set a real max length
	b, err = primitives.MarshalStringToBytes(p.APISecret, 1000)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

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
	str, newData, err := primitives.UnmarshalStringFromBytesData(newData, 1000)
	if err != nil {
		return data, err
	}
	p.APIKey = str

	str, newData, err = primitives.UnmarshalStringFromBytesData(newData, 1000)
	if err != nil {
		return data, err
	}
	p.APISecret = str
	return
}
