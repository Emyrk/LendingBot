package payment_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/src/core/payment"
)

func TestFloat64Thing(t *testing.T) {
	i, err := StringSatoshiFloatToInt64("0.01")
	if err != nil {
		t.Errorf("Error when converting 0.01: %s", err.Error())
	}
	if i != int64(SATOSHI_FLOAT*0.01) {
		t.Errorf("%d should be %d", i, int64(SATOSHI_FLOAT*0.01))
	}

	i, err = StringSatoshiFloatToInt64("-0.01")
	if err != nil {
		t.Errorf("Error when converting -0.01: %s", err.Error())
	}
	if i != int64(SATOSHI_FLOAT*-0.01) {
		t.Errorf("%d should be %d", i, int64(SATOSHI_FLOAT*-0.01))
	}

	i, err = StringSatoshiFloatToInt64("0.00000000000001")
	if err != nil {
		t.Errorf("Error when converting 0.00000000000001: %s", err.Error())
	}
	if i != 0 {
		t.Errorf("%d should be %d", i, 0)
	}

	i, err = StringSatoshiFloatToInt64("10.01")
	if err != nil {
		t.Errorf("Error when converting 10.01: %s", err.Error())
	}
	if i != int64(SATOSHI_FLOAT*10.01) {
		t.Errorf("%d should be %d", i, int64(SATOSHI_FLOAT*10.01))
	}

	_, err = StringSatoshiFloatToInt64("0.0000000000.0001")
	if err == nil {
		t.Errorf("Should have errored out with input: 0.0.0000000000.0001")
	}

	_, err = StringSatoshiFloatToInt64("0.a")
	if err == nil {
		t.Errorf("Should have errored out with input: 0.a")
	}
}

func TestInt64ToString(t *testing.T) {
	v := int64(100000000)
	s := Int64SatoshiToString(v)
	e := "1.00000000"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = 123445567
	s = Int64SatoshiToString(v)
	e = "1.23445567"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = 1
	s = Int64SatoshiToString(v)
	e = "0.00000001"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = 10
	s = Int64SatoshiToString(v)
	e = "0.00000010"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = 0
	s = Int64SatoshiToString(v)
	e = "0.00000000"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = -1
	s = Int64SatoshiToString(v)
	e = "-0.00000001"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = -10
	s = Int64SatoshiToString(v)
	e = "-0.00000010"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = -100
	s = Int64SatoshiToString(v)
	e = "-0.00000100"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = 1234455677
	s = Int64SatoshiToString(v)
	e = "12.34455677"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}

	v = -1234455677
	s = Int64SatoshiToString(v)
	e = "-12.34455677"
	if s != e {
		t.Errorf("Exp %s found %s", e, s)
	}
}
