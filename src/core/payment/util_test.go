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
