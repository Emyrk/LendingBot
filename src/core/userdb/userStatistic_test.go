package userdb_test

import (
	//"crypto"
	// "fmt"
	//"os"
	"testing"

	//"github.com/DistributedSolutions/twofactor"
	. "github.com/Emyrk/LendingBot/src/core/userdb"
)

func TestUserStat(t *testing.T) {
	u := new(UserStatistic)
	data, err := u.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	u2 := new(UserStatistic)
	data, err = u2.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	if len(data) > 0 {
		t.Error("Should be length 0")
	}
}

// func TestThisThing(t *testing.T) {
// 	thingy := func(i int, offset int) int {
// 		i += offset
// 		if i > 30 {
// 			overFlow := i - 30
// 			i = -1 + overFlow
// 		}

// 		if i < 0 {
// 			underFlow := i * -1
// 			i = 31 - underFlow
// 		}
// 		return i
// 	}

// 	for i := 0; i < 100; i++ {
// 		fmt.Println(thingy(1, -1*(i%30)))
// 	}

// }
