package userdb

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Emyrk/LendingBot/src/core/common/primitives"
)

type OldUserStatistic struct {
	Username           string    `json:"username"`
	AvailableBalance   float64   `json:"availbal"`
	ActiveLentBalance  float64   `json:"availlent"`
	OnOrderBalance     float64   `json:"onorder"`
	AverageActiveRate  float64   `json:"activerate"`
	AverageOnOrderRate float64   `json:"onorderrate"`
	Time               time.Time `json:"time"`
	Currency           string    `json:"currency"`

	TotalCurrencyMap map[string]float64

	day int
}

func (s *OldUserStatistic) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	b, err := primitives.MarshalStringToBytes(s.Username, UsernameMaxLength)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AvailableBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.ActiveLentBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.OnOrderBalance)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AverageActiveRate)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.Float64ToBytes(s.AverageOnOrderRate)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b = primitives.Uint32ToBytes(uint32(s.day))
	buf.Write(b)

	b, err = s.Time.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	b, err = primitives.MarshalStringToBytes(s.Currency, 5)
	if err != nil {
		return nil, err
	}
	buf.Write(b)

	l := len(s.TotalCurrencyMap)
	buf.Write(primitives.Uint32ToBytes(uint32(l)))

	for k, v := range s.TotalCurrencyMap {
		data, err := primitives.MarshalStringToBytes(k, 5)
		if err != nil {
			return nil, err
		}
		buf.Write(data)

		data, err = primitives.Float64ToBytes(v)
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.Next(buf.Len()), nil
}

func (s *OldUserStatistic) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}

func (s *OldUserStatistic) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[UserStatistic] A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	s.Username, newData, err = primitives.UnmarshalStringFromBytesData(newData, UsernameMaxLength)
	if err != nil {
		return nil, err
	}

	s.AvailableBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.ActiveLentBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.OnOrderBalance, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.AverageActiveRate, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	s.AverageOnOrderRate, newData, err = primitives.BytesToFloat64Data(newData)
	if err != nil {
		return nil, err
	}

	var u uint32
	u, err = primitives.BytesToUint32(newData)
	if err != nil {
		return nil, err
	}
	s.day = int(u)
	newData = newData[4:]

	td := newData[:15]
	newData = newData[15:]
	err = s.Time.UnmarshalBinary(td)
	if err != nil {
		return nil, err
	}

	s.Currency, newData, err = primitives.UnmarshalStringFromBytesData(newData, 5)
	if err != nil {
		return nil, err
	}

	l, err := primitives.BytesToUint32(newData[:4])
	if err != nil {
		return nil, err
	}
	newData = newData[4:]

	s.TotalCurrencyMap = make(map[string]float64)
	for i := 0; i < int(l); i++ {
		var key string
		key, newData, err = primitives.UnmarshalStringFromBytesData(newData, 5)
		if err != nil {
			return nil, err
		}

		var v float64
		v, newData, err = primitives.BytesToFloat64Data(newData)
		if err != nil {
			return nil, err
		}
		s.TotalCurrencyMap[key] = v
	}

	return
}
