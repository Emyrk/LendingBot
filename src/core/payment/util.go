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
	for i := 0; i < 8-len(parts[1]); i++ {
		parts[i] += "0"
	}
	return strconv.ParseInt(parts[0]+parts[1][:8], 10, 64)
}
