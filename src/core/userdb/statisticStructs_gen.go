package userdb

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *AllLendingHistoryEntry) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "PoloSet":
			z.PoloSet, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "BitfinSet":
			z.BitfinSet, err = dc.ReadBool()
			if err != nil {
				return
			}
		case "PoloniexData":
			var zwht uint32
			zwht, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.PoloniexData == nil && zwht > 0 {
				z.PoloniexData = make(map[string]*LendingHistoryEntry, zwht)
			} else if len(z.PoloniexData) > 0 {
				for key, _ := range z.PoloniexData {
					delete(z.PoloniexData, key)
				}
			}
			for zwht > 0 {
				zwht--
				var zxvk string
				var zbzg *LendingHistoryEntry
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
						zbzg = new(LendingHistoryEntry)
					}
					err = zbzg.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.PoloniexData[zxvk] = zbzg
			}
		case "BitfinexData":
			var zhct uint32
			zhct, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.BitfinexData == nil && zhct > 0 {
				z.BitfinexData = make(map[string]*LendingHistoryEntry, zhct)
			} else if len(z.BitfinexData) > 0 {
				for key, _ := range z.BitfinexData {
					delete(z.BitfinexData, key)
				}
			}
			for zhct > 0 {
				zhct--
				var zbai string
				var zcmr *LendingHistoryEntry
				zbai, err = dc.ReadString()
				if err != nil {
					return
				}
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					zcmr = nil
				} else {
					if zcmr == nil {
						zcmr = new(LendingHistoryEntry)
					}
					err = zcmr.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.BitfinexData[zbai] = zcmr
			}
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "ShortTime":
			z.ShortTime, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Username":
			z.Username, err = dc.ReadString()
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
func (z *AllLendingHistoryEntry) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "PoloSet"
	err = en.Append(0x87, 0xa7, 0x50, 0x6f, 0x6c, 0x6f, 0x53, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.PoloSet)
	if err != nil {
		return
	}
	// write "BitfinSet"
	err = en.Append(0xa9, 0x42, 0x69, 0x74, 0x66, 0x69, 0x6e, 0x53, 0x65, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteBool(z.BitfinSet)
	if err != nil {
		return
	}
	// write "PoloniexData"
	err = en.Append(0xac, 0x50, 0x6f, 0x6c, 0x6f, 0x6e, 0x69, 0x65, 0x78, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.PoloniexData)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.PoloniexData {
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
	// write "BitfinexData"
	err = en.Append(0xac, 0x42, 0x69, 0x74, 0x66, 0x69, 0x6e, 0x65, 0x78, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.BitfinexData)))
	if err != nil {
		return
	}
	for zbai, zcmr := range z.BitfinexData {
		err = en.WriteString(zbai)
		if err != nil {
			return
		}
		if zcmr == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zcmr.EncodeMsg(en)
			if err != nil {
				return
			}
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
	// write "ShortTime"
	err = en.Append(0xa9, 0x53, 0x68, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ShortTime)
	if err != nil {
		return
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
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AllLendingHistoryEntry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "PoloSet"
	o = append(o, 0x87, 0xa7, 0x50, 0x6f, 0x6c, 0x6f, 0x53, 0x65, 0x74)
	o = msgp.AppendBool(o, z.PoloSet)
	// string "BitfinSet"
	o = append(o, 0xa9, 0x42, 0x69, 0x74, 0x66, 0x69, 0x6e, 0x53, 0x65, 0x74)
	o = msgp.AppendBool(o, z.BitfinSet)
	// string "PoloniexData"
	o = append(o, 0xac, 0x50, 0x6f, 0x6c, 0x6f, 0x6e, 0x69, 0x65, 0x78, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.PoloniexData)))
	for zxvk, zbzg := range z.PoloniexData {
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
	// string "BitfinexData"
	o = append(o, 0xac, 0x42, 0x69, 0x74, 0x66, 0x69, 0x6e, 0x65, 0x78, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.BitfinexData)))
	for zbai, zcmr := range z.BitfinexData {
		o = msgp.AppendString(o, zbai)
		if zcmr == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zcmr.MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	// string "Time"
	o = append(o, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	// string "ShortTime"
	o = append(o, 0xa9, 0x53, 0x68, 0x6f, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendString(o, z.ShortTime)
	// string "Username"
	o = append(o, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Username)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AllLendingHistoryEntry) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "PoloSet":
			z.PoloSet, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "BitfinSet":
			z.BitfinSet, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				return
			}
		case "PoloniexData":
			var zxhx uint32
			zxhx, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.PoloniexData == nil && zxhx > 0 {
				z.PoloniexData = make(map[string]*LendingHistoryEntry, zxhx)
			} else if len(z.PoloniexData) > 0 {
				for key, _ := range z.PoloniexData {
					delete(z.PoloniexData, key)
				}
			}
			for zxhx > 0 {
				var zxvk string
				var zbzg *LendingHistoryEntry
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
						zbzg = new(LendingHistoryEntry)
					}
					bts, err = zbzg.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.PoloniexData[zxvk] = zbzg
			}
		case "BitfinexData":
			var zlqf uint32
			zlqf, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.BitfinexData == nil && zlqf > 0 {
				z.BitfinexData = make(map[string]*LendingHistoryEntry, zlqf)
			} else if len(z.BitfinexData) > 0 {
				for key, _ := range z.BitfinexData {
					delete(z.BitfinexData, key)
				}
			}
			for zlqf > 0 {
				var zbai string
				var zcmr *LendingHistoryEntry
				zlqf--
				zbai, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					zcmr = nil
				} else {
					if zcmr == nil {
						zcmr = new(LendingHistoryEntry)
					}
					bts, err = zcmr.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.BitfinexData[zbai] = zcmr
			}
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "ShortTime":
			z.ShortTime, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
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
func (z *AllLendingHistoryEntry) Msgsize() (s int) {
	s = 1 + 8 + msgp.BoolSize + 10 + msgp.BoolSize + 13 + msgp.MapHeaderSize
	if z.PoloniexData != nil {
		for zxvk, zbzg := range z.PoloniexData {
			_ = zbzg
			s += msgp.StringPrefixSize + len(zxvk)
			if zbzg == nil {
				s += msgp.NilSize
			} else {
				s += zbzg.Msgsize()
			}
		}
	}
	s += 13 + msgp.MapHeaderSize
	if z.BitfinexData != nil {
		for zbai, zcmr := range z.BitfinexData {
			_ = zcmr
			s += msgp.StringPrefixSize + len(zbai)
			if zcmr == nil {
				s += msgp.NilSize
			} else {
				s += zcmr.Msgsize()
			}
		}
	}
	s += 5 + msgp.TimeSize + 10 + msgp.StringPrefixSize + len(z.ShortTime) + 9 + msgp.StringPrefixSize + len(z.Username)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AllUserStatistic) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zeff uint32
	zeff, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zeff > 0 {
		zeff--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Currencies":
			var zrsw uint32
			zrsw, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Currencies == nil && zrsw > 0 {
				z.Currencies = make(map[string]*UserStatistic, zrsw)
			} else if len(z.Currencies) > 0 {
				for key, _ := range z.Currencies {
					delete(z.Currencies, key)
				}
			}
			for zrsw > 0 {
				zrsw--
				var zdaf string
				var zpks *UserStatistic
				zdaf, err = dc.ReadString()
				if err != nil {
					return
				}
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					zpks = nil
				} else {
					if zpks == nil {
						zpks = new(UserStatistic)
					}
					err = zpks.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.Currencies[zdaf] = zpks
			}
		case "Username":
			z.Username, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zxpk uint32
			zxpk, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.TotalCurrencyMap == nil && zxpk > 0 {
				z.TotalCurrencyMap = make(map[string]float64, zxpk)
			} else if len(z.TotalCurrencyMap) > 0 {
				for key, _ := range z.TotalCurrencyMap {
					delete(z.TotalCurrencyMap, key)
				}
			}
			for zxpk > 0 {
				zxpk--
				var zjfb string
				var zcxo float64
				zjfb, err = dc.ReadString()
				if err != nil {
					return
				}
				zcxo, err = dc.ReadFloat64()
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zjfb] = zcxo
			}
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "Exchange":
			{
				var zdnj string
				zdnj, err = dc.ReadString()
				z.Exchange = UserExchange(zdnj)
			}
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
	// map header, size 5
	// write "Currencies"
	err = en.Append(0x85, 0xaa, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Currencies)))
	if err != nil {
		return
	}
	for zdaf, zpks := range z.Currencies {
		err = en.WriteString(zdaf)
		if err != nil {
			return
		}
		if zpks == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zpks.EncodeMsg(en)
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
	for zjfb, zcxo := range z.TotalCurrencyMap {
		err = en.WriteString(zjfb)
		if err != nil {
			return
		}
		err = en.WriteFloat64(zcxo)
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
	// write "Exchange"
	err = en.Append(0xa8, 0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(string(z.Exchange))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AllUserStatistic) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Currencies"
	o = append(o, 0x85, 0xaa, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Currencies)))
	for zdaf, zpks := range z.Currencies {
		o = msgp.AppendString(o, zdaf)
		if zpks == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zpks.MarshalMsg(o)
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
	for zjfb, zcxo := range z.TotalCurrencyMap {
		o = msgp.AppendString(o, zjfb)
		o = msgp.AppendFloat64(o, zcxo)
	}
	// string "Time"
	o = append(o, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	// string "Exchange"
	o = append(o, 0xa8, 0x45, 0x78, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65)
	o = msgp.AppendString(o, string(z.Exchange))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AllUserStatistic) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zobc uint32
	zobc, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zobc > 0 {
		zobc--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Currencies":
			var zsnv uint32
			zsnv, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Currencies == nil && zsnv > 0 {
				z.Currencies = make(map[string]*UserStatistic, zsnv)
			} else if len(z.Currencies) > 0 {
				for key, _ := range z.Currencies {
					delete(z.Currencies, key)
				}
			}
			for zsnv > 0 {
				var zdaf string
				var zpks *UserStatistic
				zsnv--
				zdaf, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					zpks = nil
				} else {
					if zpks == nil {
						zpks = new(UserStatistic)
					}
					bts, err = zpks.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.Currencies[zdaf] = zpks
			}
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zkgt uint32
			zkgt, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.TotalCurrencyMap == nil && zkgt > 0 {
				z.TotalCurrencyMap = make(map[string]float64, zkgt)
			} else if len(z.TotalCurrencyMap) > 0 {
				for key, _ := range z.TotalCurrencyMap {
					delete(z.TotalCurrencyMap, key)
				}
			}
			for zkgt > 0 {
				var zjfb string
				var zcxo float64
				zkgt--
				zjfb, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				zcxo, bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zjfb] = zcxo
			}
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "Exchange":
			{
				var zema string
				zema, bts, err = msgp.ReadStringBytes(bts)
				z.Exchange = UserExchange(zema)
			}
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
		for zdaf, zpks := range z.Currencies {
			_ = zpks
			s += msgp.StringPrefixSize + len(zdaf)
			if zpks == nil {
				s += msgp.NilSize
			} else {
				s += zpks.Msgsize()
			}
		}
	}
	s += 9 + msgp.StringPrefixSize + len(z.Username) + 17 + msgp.MapHeaderSize
	if z.TotalCurrencyMap != nil {
		for zjfb, zcxo := range z.TotalCurrencyMap {
			_ = zcxo
			s += msgp.StringPrefixSize + len(zjfb) + msgp.Float64Size
		}
	}
	s += 5 + msgp.TimeSize + 9 + msgp.StringPrefixSize + len(string(z.Exchange))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *LendingHistoryEntry) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zpez uint32
	zpez, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zpez > 0 {
		zpez--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Earned":
			z.Earned, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "Fees":
			z.Fees, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "AvgDuration":
			z.AvgDuration, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, err = dc.ReadString()
			if err != nil {
				return
			}
		case "LoanCounts":
			z.LoanCounts, err = dc.ReadInt()
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
func (z *LendingHistoryEntry) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "Earned"
	err = en.Append(0x85, 0xa6, 0x45, 0x61, 0x72, 0x6e, 0x65, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.Earned)
	if err != nil {
		return
	}
	// write "Fees"
	err = en.Append(0xa4, 0x46, 0x65, 0x65, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.Fees)
	if err != nil {
		return
	}
	// write "AvgDuration"
	err = en.Append(0xab, 0x41, 0x76, 0x67, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.AvgDuration)
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
	// write "LoanCounts"
	err = en.Append(0xaa, 0x4c, 0x6f, 0x61, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.LoanCounts)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *LendingHistoryEntry) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Earned"
	o = append(o, 0x85, 0xa6, 0x45, 0x61, 0x72, 0x6e, 0x65, 0x64)
	o = msgp.AppendFloat64(o, z.Earned)
	// string "Fees"
	o = append(o, 0xa4, 0x46, 0x65, 0x65, 0x73)
	o = msgp.AppendFloat64(o, z.Fees)
	// string "AvgDuration"
	o = append(o, 0xab, 0x41, 0x76, 0x67, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	o = msgp.AppendFloat64(o, z.AvgDuration)
	// string "Currency"
	o = append(o, 0xa8, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79)
	o = msgp.AppendString(o, z.Currency)
	// string "LoanCounts"
	o = append(o, 0xaa, 0x4c, 0x6f, 0x61, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x73)
	o = msgp.AppendInt(o, z.LoanCounts)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *LendingHistoryEntry) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zqke uint32
	zqke, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zqke > 0 {
		zqke--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Earned":
			z.Earned, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "Fees":
			z.Fees, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "AvgDuration":
			z.AvgDuration, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "LoanCounts":
			z.LoanCounts, bts, err = msgp.ReadIntBytes(bts)
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
func (z *LendingHistoryEntry) Msgsize() (s int) {
	s = 1 + 7 + msgp.Float64Size + 5 + msgp.Float64Size + 12 + msgp.Float64Size + 9 + msgp.StringPrefixSize + len(z.Currency) + 11 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *MongoAllUserStatistics) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zyzr uint32
	zyzr, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zyzr > 0 {
		zyzr--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Username":
			z.Username, err = dc.ReadString()
			if err != nil {
				return
			}
		case "UserStatistics":
			var zywj uint32
			zywj, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.UserStatistics) >= int(zywj) {
				z.UserStatistics = (z.UserStatistics)[:zywj]
			} else {
				z.UserStatistics = make([]AllUserStatistic, zywj)
			}
			for zqyh := range z.UserStatistics {
				err = z.UserStatistics[zqyh].DecodeMsg(dc)
				if err != nil {
					return
				}
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
func (z *MongoAllUserStatistics) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "Username"
	err = en.Append(0x82, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Username)
	if err != nil {
		return
	}
	// write "UserStatistics"
	err = en.Append(0xae, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.UserStatistics)))
	if err != nil {
		return
	}
	for zqyh := range z.UserStatistics {
		err = z.UserStatistics[zqyh].EncodeMsg(en)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *MongoAllUserStatistics) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Username"
	o = append(o, 0x82, 0xa8, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Username)
	// string "UserStatistics"
	o = append(o, 0xae, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x69, 0x73, 0x74, 0x69, 0x63, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.UserStatistics)))
	for zqyh := range z.UserStatistics {
		o, err = z.UserStatistics[zqyh].MarshalMsg(o)
		if err != nil {
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *MongoAllUserStatistics) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zjpj uint32
	zjpj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zjpj > 0 {
		zjpj--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "UserStatistics":
			var zzpf uint32
			zzpf, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.UserStatistics) >= int(zzpf) {
				z.UserStatistics = (z.UserStatistics)[:zzpf]
			} else {
				z.UserStatistics = make([]AllUserStatistic, zzpf)
			}
			for zqyh := range z.UserStatistics {
				bts, err = z.UserStatistics[zqyh].UnmarshalMsg(bts)
				if err != nil {
					return
				}
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
func (z *MongoAllUserStatistics) Msgsize() (s int) {
	s = 1 + 9 + msgp.StringPrefixSize + len(z.Username) + 15 + msgp.ArrayHeaderSize
	for zqyh := range z.UserStatistics {
		s += z.UserStatistics[zqyh].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PoloniexStat) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zrfe uint32
	zrfe, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zrfe > 0 {
		zrfe--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "Rate":
			z.Rate, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, err = dc.ReadString()
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
func (z PoloniexStat) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Time"
	err = en.Append(0x83, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteTime(z.Time)
	if err != nil {
		return
	}
	// write "Rate"
	err = en.Append(0xa4, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.Rate)
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
	return
}

// MarshalMsg implements msgp.Marshaler
func (z PoloniexStat) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Time"
	o = append(o, 0x83, 0xa4, 0x54, 0x69, 0x6d, 0x65)
	o = msgp.AppendTime(o, z.Time)
	// string "Rate"
	o = append(o, 0xa4, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.Rate)
	// string "Currency"
	o = append(o, 0xa8, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79)
	o = msgp.AppendString(o, z.Currency)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *PoloniexStat) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zgmo uint32
	zgmo, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zgmo > 0 {
		zgmo--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "Rate":
			z.Rate, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "Currency":
			z.Currency, bts, err = msgp.ReadStringBytes(bts)
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
func (z PoloniexStat) Msgsize() (s int) {
	s = 1 + 5 + msgp.TimeSize + 5 + msgp.Float64Size + 9 + msgp.StringPrefixSize + len(z.Currency)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserExchange) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var ztaf string
		ztaf, err = dc.ReadString()
		(*z) = UserExchange(ztaf)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UserExchange) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UserExchange) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UserExchange) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zeth string
		zeth, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UserExchange(zeth)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UserExchange) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UserStatistic) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zsbz uint32
	zsbz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zsbz > 0 {
		zsbz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BTCRate":
			z.BTCRate, err = dc.ReadFloat64()
			if err != nil {
				return
			}
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
		case "HighestRate":
			z.HighestRate, err = dc.ReadFloat64()
			if err != nil {
				return
			}
		case "LowestRate":
			z.LowestRate, err = dc.ReadFloat64()
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
	// map header, size 10
	// write "BTCRate"
	err = en.Append(0x8a, 0xa7, 0x42, 0x54, 0x43, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.BTCRate)
	if err != nil {
		return
	}
	// write "AvailableBalance"
	err = en.Append(0xb0, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
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
	// write "HighestRate"
	err = en.Append(0xab, 0x48, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.HighestRate)
	if err != nil {
		return
	}
	// write "LowestRate"
	err = en.Append(0xaa, 0x4c, 0x6f, 0x77, 0x65, 0x73, 0x74, 0x52, 0x61, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteFloat64(z.LowestRate)
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
	// map header, size 10
	// string "BTCRate"
	o = append(o, 0x8a, 0xa7, 0x42, 0x54, 0x43, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.BTCRate)
	// string "AvailableBalance"
	o = append(o, 0xb0, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
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
	// string "HighestRate"
	o = append(o, 0xab, 0x48, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.HighestRate)
	// string "LowestRate"
	o = append(o, 0xaa, 0x4c, 0x6f, 0x77, 0x65, 0x73, 0x74, 0x52, 0x61, 0x74, 0x65)
	o = msgp.AppendFloat64(o, z.LowestRate)
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
	var zrjx uint32
	zrjx, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zrjx > 0 {
		zrjx--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "BTCRate":
			z.BTCRate, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
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
		case "HighestRate":
			z.HighestRate, bts, err = msgp.ReadFloat64Bytes(bts)
			if err != nil {
				return
			}
		case "LowestRate":
			z.LowestRate, bts, err = msgp.ReadFloat64Bytes(bts)
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
	s = 1 + 8 + msgp.Float64Size + 17 + msgp.Float64Size + 18 + msgp.Float64Size + 15 + msgp.Float64Size + 18 + msgp.Float64Size + 19 + msgp.Float64Size + 12 + msgp.Float64Size + 11 + msgp.Float64Size + 9 + msgp.StringPrefixSize + len(z.Currency) + 5 + msgp.TimeSize
	return
}
