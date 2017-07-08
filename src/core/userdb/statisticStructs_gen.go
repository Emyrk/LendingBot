package userdb

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *AllLendingHistoryEntry) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zbai uint32
	zbai, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zbai > 0 {
		zbai--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Data":
			var zcmr uint32
			zcmr, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Data == nil && zcmr > 0 {
				z.Data = make(map[string]*LendingHistoryEntry, zcmr)
			} else if len(z.Data) > 0 {
				for key, _ := range z.Data {
					delete(z.Data, key)
				}
			}
			for zcmr > 0 {
				zcmr--
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
				z.Data[zxvk] = zbzg
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
	// map header, size 4
	// write "Data"
	err = en.Append(0x84, 0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Data)))
	if err != nil {
		return
	}
	for zxvk, zbzg := range z.Data {
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
	// map header, size 4
	// string "Data"
	o = append(o, 0x84, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendMapHeader(o, uint32(len(z.Data)))
	for zxvk, zbzg := range z.Data {
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
	var zajw uint32
	zajw, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Data":
			var zwht uint32
			zwht, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Data == nil && zwht > 0 {
				z.Data = make(map[string]*LendingHistoryEntry, zwht)
			} else if len(z.Data) > 0 {
				for key, _ := range z.Data {
					delete(z.Data, key)
				}
			}
			for zwht > 0 {
				var zxvk string
				var zbzg *LendingHistoryEntry
				zwht--
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
				z.Data[zxvk] = zbzg
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
	s = 1 + 5 + msgp.MapHeaderSize
	if z.Data != nil {
		for zxvk, zbzg := range z.Data {
			_ = zbzg
			s += msgp.StringPrefixSize + len(zxvk)
			if zbzg == nil {
				s += msgp.NilSize
			} else {
				s += zbzg.Msgsize()
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
		case "Currencies":
			var zpks uint32
			zpks, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Currencies == nil && zpks > 0 {
				z.Currencies = make(map[string]*UserStatistic, zpks)
			} else if len(z.Currencies) > 0 {
				for key, _ := range z.Currencies {
					delete(z.Currencies, key)
				}
			}
			for zpks > 0 {
				zpks--
				var zhct string
				var zcua *UserStatistic
				zhct, err = dc.ReadString()
				if err != nil {
					return
				}
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					zcua = nil
				} else {
					if zcua == nil {
						zcua = new(UserStatistic)
					}
					err = zcua.DecodeMsg(dc)
					if err != nil {
						return
					}
				}
				z.Currencies[zhct] = zcua
			}
		case "Username":
			z.Username, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zjfb uint32
			zjfb, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.TotalCurrencyMap == nil && zjfb > 0 {
				z.TotalCurrencyMap = make(map[string]float64, zjfb)
			} else if len(z.TotalCurrencyMap) > 0 {
				for key, _ := range z.TotalCurrencyMap {
					delete(z.TotalCurrencyMap, key)
				}
			}
			for zjfb > 0 {
				zjfb--
				var zxhx string
				var zlqf float64
				zxhx, err = dc.ReadString()
				if err != nil {
					return
				}
				zlqf, err = dc.ReadFloat64()
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zxhx] = zlqf
			}
		case "Time":
			z.Time, err = dc.ReadTime()
			if err != nil {
				return
			}
		case "Exchange":
			{
				var zcxo string
				zcxo, err = dc.ReadString()
				z.Exchange = UserExchange(zcxo)
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
	for zhct, zcua := range z.Currencies {
		err = en.WriteString(zhct)
		if err != nil {
			return
		}
		if zcua == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = zcua.EncodeMsg(en)
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
	for zxhx, zlqf := range z.TotalCurrencyMap {
		err = en.WriteString(zxhx)
		if err != nil {
			return
		}
		err = en.WriteFloat64(zlqf)
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
	for zhct, zcua := range z.Currencies {
		o = msgp.AppendString(o, zhct)
		if zcua == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = zcua.MarshalMsg(o)
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
	for zxhx, zlqf := range z.TotalCurrencyMap {
		o = msgp.AppendString(o, zxhx)
		o = msgp.AppendFloat64(o, zlqf)
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
	var zeff uint32
	zeff, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zeff > 0 {
		zeff--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Currencies":
			var zrsw uint32
			zrsw, bts, err = msgp.ReadMapHeaderBytes(bts)
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
				var zhct string
				var zcua *UserStatistic
				zrsw--
				zhct, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					zcua = nil
				} else {
					if zcua == nil {
						zcua = new(UserStatistic)
					}
					bts, err = zcua.UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
				z.Currencies[zhct] = zcua
			}
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TotalCurrencyMap":
			var zxpk uint32
			zxpk, bts, err = msgp.ReadMapHeaderBytes(bts)
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
				var zxhx string
				var zlqf float64
				zxpk--
				zxhx, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				zlqf, bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
				z.TotalCurrencyMap[zxhx] = zlqf
			}
		case "Time":
			z.Time, bts, err = msgp.ReadTimeBytes(bts)
			if err != nil {
				return
			}
		case "Exchange":
			{
				var zdnj string
				zdnj, bts, err = msgp.ReadStringBytes(bts)
				z.Exchange = UserExchange(zdnj)
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
		for zhct, zcua := range z.Currencies {
			_ = zcua
			s += msgp.StringPrefixSize + len(zhct)
			if zcua == nil {
				s += msgp.NilSize
			} else {
				s += zcua.Msgsize()
			}
		}
	}
	s += 9 + msgp.StringPrefixSize + len(z.Username) + 17 + msgp.MapHeaderSize
	if z.TotalCurrencyMap != nil {
		for zxhx, zlqf := range z.TotalCurrencyMap {
			_ = zlqf
			s += msgp.StringPrefixSize + len(zxhx) + msgp.Float64Size
		}
	}
	s += 5 + msgp.TimeSize + 9 + msgp.StringPrefixSize + len(string(z.Exchange))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *LendingHistoryEntry) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zobc uint32
	zobc, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zobc > 0 {
		zobc--
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
	var zsnv uint32
	zsnv, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zsnv > 0 {
		zsnv--
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
	var zema uint32
	zema, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zema > 0 {
		zema--
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
			var zpez uint32
			zpez, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.UserStatistics) >= int(zpez) {
				z.UserStatistics = (z.UserStatistics)[:zpez]
			} else {
				z.UserStatistics = make([]AllUserStatistic, zpez)
			}
			for zkgt := range z.UserStatistics {
				err = z.UserStatistics[zkgt].DecodeMsg(dc)
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
	for zkgt := range z.UserStatistics {
		err = z.UserStatistics[zkgt].EncodeMsg(en)
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
	for zkgt := range z.UserStatistics {
		o, err = z.UserStatistics[zkgt].MarshalMsg(o)
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
		case "Username":
			z.Username, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "UserStatistics":
			var zqyh uint32
			zqyh, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.UserStatistics) >= int(zqyh) {
				z.UserStatistics = (z.UserStatistics)[:zqyh]
			} else {
				z.UserStatistics = make([]AllUserStatistic, zqyh)
			}
			for zkgt := range z.UserStatistics {
				bts, err = z.UserStatistics[zkgt].UnmarshalMsg(bts)
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
	for zkgt := range z.UserStatistics {
		s += z.UserStatistics[zkgt].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *PoloniexStat) DecodeMsg(dc *msgp.Reader) (err error) {
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
	var zywj uint32
	zywj, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zywj > 0 {
		zywj--
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
		var zjpj string
		zjpj, err = dc.ReadString()
		(*z) = UserExchange(zjpj)
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
		var zzpf string
		zzpf, bts, err = msgp.ReadStringBytes(bts)
		(*z) = UserExchange(zzpf)
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
