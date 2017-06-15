package userdb

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *AllUserStatistic) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zajw uint32
	zajw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Currencies":
			var zwht uint32
			zwht, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Currencies == nil && zwht > 0 {
				z.Currencies = make(map[string]*UserStatistic, zwht)
			} else if len(z.Currencies) > 0 {
				for key, _ := range z.Currencies {
					delete(z.Currencies, key)
				}
			}
			for zwht > 0 {
				zwht--
				var zxvk string
				var zbzg *UserStatistic
				zxvk, err = dc.ReadString()
				if err != nil {
					return
				}
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					zbzg = nil
				} else {
					if zbzg == nil {
						zbzg = new(UserStatistic)
					}
					err = zbzg.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.Currencies[zxvk] = zbzg
			}
		case "Username":
			z.Username, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zhct uint32
			zhct, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.TotalCurrencyMap == nil && zhct > 0 {
				z.TotalCurrencyMap = make(map[string]float64, zhct)
			} else if len(z.TotalCurrencyMap) > 0 {
				for key, _ := range z.TotalCurrencyMap {
					delete(z.TotalCurrencyMap, key)
				}
			}
			for zhct > 0 {
				zhct--
				var zbai string
				var zcmr float64
				zbai, err = dc.ReadString()
				if err != nil {
					return
				}
				zcmr, err = dc.ReadFloat64()
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zbai] = zcmr
			}
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *AllUserStatistic) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "Currencies"
	err = en.Append(0x84, 0xaa, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Currencies)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.Currencies {
		err = en.WriteString(zxvk)
		if err != nil {
			return
		}
		if zbzg == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zbzg.EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	// write "Username"
	err = en.Append(0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Username)
	if err != nil {
		return
	}
	// write "TotalCurrencyMap"
	err = en.Append(0xb0, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x4d, 0x61, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.TotalCurrencyMap)))
	if err != nil {
		return
	}
	for zbai, zcmr := range z.TotalCurrencyMap {
		err = en.WriteString(zbai)
		if err != nil {
			return
		}
		err = en.WriteFloat64(zcmr)
		if err != nil {
			return
		}
	}
	// write "Time"
	err = en.Append(0xa4, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AllUserStatistic) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "Currencies"
	o = append(o, 0x84, 0xaa, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Currencies)))
	for zxvk, zbzg := range z.Currencies {
		o = msgp.AppendString(o, zxvk)
		if zbzg == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zbzg.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "Username"
	o = append(o, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Username)
	// string "TotalCurrencyMap"
	o = append(o, 0xb0, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x4d, 0x61, 0x70)
	o = msgp.AppendMapHeader(o, uint32(len(z.TotalCurrencyMap)))
	for zbai, zcmr := range z.TotalCurrencyMap {
		o = msgp.AppendString(o, zbai)
		o = msgp.AppendFloat64(o, zcmr)
	}
	// string "Time"
	o = append(o, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AllUserStatistic) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zcua uint32
	zcua, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zcua > 0 {
		zcua--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Currencies":
			var zxhx uint32
			zxhx, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Currencies == nil && zxhx > 0 {
				z.Currencies = make(map[string]*UserStatistic, zxhx)
			} else if len(z.Currencies) > 0 {
				for key, _ := range z.Currencies {
					delete(z.Currencies, key)
				}
			}
			for zxhx > 0 {
				var zxvk string
				var zbzg *UserStatistic
				zxhx--
				zxvk, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					zbzg = nil
				} else {
					if zbzg == nil {
						zbzg = new(UserStatistic)
					}
					bts, err = zbzg.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.Currencies[zxvk] = zbzg
			}
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zlqf uint32
			zlqf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.TotalCurrencyMap == nil && zlqf > 0 {
				z.TotalCurrencyMap = make(map[string]float64, zlqf)
			} else if len(z.TotalCurrencyMap) > 0 {
				for key, _ := range z.TotalCurrencyMap {
					delete(z.TotalCurrencyMap, key)
				}
			}
			for zlqf > 0 {
				var zbai string
				var zcmr float64
				zlqf--
				zbai, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				zcmr, bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zbai] = zcmr
			}
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *AllUserStatistic) Msgsize() (s int) {
	s = 1 + 11 + msgp.MapHeaderSize
	if z.Currencies != nil {
		for zxvk, zbzg := range z.Currencies {
			_ = zbzg
			s += msgp.StringPrefixSize + len(zxvk)
			if zbzg == nil {
				s += msgp.NilSize
			} else {
				s += zbzg.Msgsize()
			}
		}
	}
	s += 9 + msgp.StringPrefixSize + len(z.Username) + 17 + msgp.MapHeaderSize
	if z.TotalCurrencyMap != nil {
		for zbai, zcmr := range z.TotalCurrencyMap {
			_ = zcmr
			s += msgp.StringPrefixSize + len(zbai) + msgp.Float64Size
		}
	}
	s += 5 + msgp.TimeSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserStatistic) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zdaf uint32
	zdaf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zdaf > 0 {
		zdaf--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AvailableBalance":
			z.AvailableBalance, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "ActiveLentBalance":
			z.ActiveLentBalance, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "OnOrderBalance":
			z.OnOrderBalance, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "AverageActiveRate":
			z.AverageActiveRate, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "AverageOnOrderRate":
			z.AverageOnOrderRate, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *UserStatistic) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "AvailableBalance"
	err = en.Append(0x87, 0xb0, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.AvailableBalance)
	if err != nil {
		return
	}
	// write "ActiveLentBalance"
	err = en.Append(0xb1, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x4c, 0x65, 0x6e, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.ActiveLentBalance)
	if err != nil {
		return
	}
	// write "OnOrderBalance"
	err = en.Append(0xae, 0x4f, 0x6e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.OnOrderBalance)
	if err != nil {
		return
	}
	// write "AverageActiveRate"
	err = en.Append(0xb1, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.AverageActiveRate)
	if err != nil {
		return
	}
	// write "AverageOnOrderRate"
	err = en.Append(0xb2, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x4f, 0x6e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.AverageOnOrderRate)
	if err != nil {
		return
	}
	// write "Currency"
	err = en.Append(0xa8, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Currency)
	if err != nil {
		return
	}
	// write "Time"
	err = en.Append(0xa4, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *UserStatistic) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "AvailableBalance"
	o = append(o, 0x87, 0xb0, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o = msgp.AppendFloat64(o, z.AvailableBalance)
	// string "ActiveLentBalance"
	o = append(o, 0xb1, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x4c, 0x65, 0x6e, 0x74, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o = msgp.AppendFloat64(o, z.ActiveLentBalance)
	// string "OnOrderBalance"
	o = append(o, 0xae, 0x4f, 0x6e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o = msgp.AppendFloat64(o, z.OnOrderBalance)
	// string "AverageActiveRate"
	o = append(o, 0xb1, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.AverageActiveRate)
	// string "AverageOnOrderRate"
	o = append(o, 0xb2, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x4f, 0x6e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.AverageOnOrderRate)
	// string "Currency"
	o = append(o, 0xa8, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79)
	o = msgp.AppendString(o, z.Currency)
	// string "Time"
	o = append(o, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserStatistic) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zpks uint32
	zpks, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zpks > 0 {
		zpks--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AvailableBalance":
			z.AvailableBalance, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "ActiveLentBalance":
			z.ActiveLentBalance, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "OnOrderBalance":
			z.OnOrderBalance, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "AverageActiveRate":
			z.AverageActiveRate, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "AverageOnOrderRate":
			z.AverageOnOrderRate, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *UserStatistic) Msgsize() (s int) {
	s = 1 + 17 + msgp.Float64Size + 18 + msgp.Float64Size + 15 + msgp.Float64Size + 18 + msgp.Float64Size + 19 + msgp.Float64Size + 9 + msgp.StringPrefixSize + len(z.Currency) + 5 + msgp.TimeSize
	return
}
