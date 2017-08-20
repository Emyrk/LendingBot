package payment

import (
	"fmt"
	"strconv"
	"strings"
)

func StringSatoshiFloatToInt64(str string) (int64, error) {
	parts := strings.Split(str, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("Invalid number: %s", str)
	}
	ap := 8 - len(parts[1])
	for i := 0; i < ap; i++ {
		parts[1] += "0"
	}
	return strconv.ParseInt(parts[0]+parts[1][:8], 10, 64)
}

func Int64SatoshiToString(val int64) string {
	neg := ""
	if val < 0 {
		val = val * -1
		neg = "-"
	}
	str := fmt.Sprintf("%d", val)
	if len(str) > 8 { // Place the decimal place where it needs to go
		end := str[len(str)-8:]
		beg := str[:len(str)-8]
		return neg + beg + "." + end
	}

	for i := len(str); i < 8; i++ {
		str = "0" + str
	}
	return neg + "0." + str
}

func RoundFloat(f float64, sigFigs int32) (float64, error) {
	roundTo := fmt.Sprintf("%%.%df", DEFAULT_REDUCTION_ROUND)
	rf, err := strconv.ParseFloat(fmt.Sprintf(roundTo, f), 64)
	if err != nil {
		return 0.0, err
	}
	return rf, nil
}
