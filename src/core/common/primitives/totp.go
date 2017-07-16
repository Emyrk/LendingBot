package primitives

import (
	"github.com/DistributedSolutions/twofactor"
	"gopkg.in/mgo.v2/bson"
	// "fmt"
)

type Totp struct {
	*twofactor.Totp
}

// GetBSON implements bson.Getter.
func (a *Totp) GetBSON() (interface{}, error) {
	b, err := a.ToBytes()
	if err != nil {
		return nil, err
	}
	return struct {
		Totp []byte `json:"totp" bson:"totp"`
	}{
		Totp: b,
	}, nil
}

// SetBSON implements bson.Setter.
func (a *Totp) SetBSON(raw bson.Raw) error {
	if len(raw.Data) > 0 {
		decoded := new(struct {
			Totp []byte `json:"totp" bson:"totp"`
		})
		err := raw.Unmarshal(decoded)
		if err != nil {
			return err
		}
		newTotp, err := twofactor.TOTPFromBytes(decoded.Totp, "HodlZone")
		if err != nil {
			return err
		}
		a.Totp = newTotp
	}
	return nil
}
