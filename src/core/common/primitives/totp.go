package primitives

import (
	"fmt"

	"github.com/DistributedSolutions/twofactor"
	"gopkg.in/mgo.v2/bson"
)

type Totp struct {
	*twofactor.Totp
}

// GetBSON implements bson.Getter.
func (a *Totp) GetBSON() (interface{}, error) {
	if a.Totp != nil {
		b, err := a.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("Error marshalling totp bson [%s]: %s", a, err.Error())
		}
		return struct {
			Totp []byte `json:"totp" bson:"totp"`
		}{
			Totp: b,
		}, nil
	}
	return nil, nil
}

// SetBSON implements bson.Setter.
func (a *Totp) SetBSON(raw bson.Raw) error {
	if len(raw.Data) > 0 {
		decoded := new(struct {
			Totp []byte `json:"totp" bson:"totp"`
		})
		err := raw.Unmarshal(decoded)
		if err != nil {
			return fmt.Errorf("Error unmarshalling raw code: %s", err.Error())
		}
		newTotp, err := twofactor.TOTPFromBytes(decoded.Totp, "HodlZone")
		if err != nil {
			return fmt.Errorf("Error getting totp from bytes: %s", err.Error())
		}
		a.Totp = newTotp
	}
	return nil
}
