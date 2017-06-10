package queuer

import (
	"testing"

	"github.com/Emyrk/LendingBot/src/core"
	"github.com/Emyrk/LendingBot/src/lender"
	. "github.com/Emyrk/LendingBot/src/queuer"
)

func main() {
	s := core.NewStateWithMap()
	l := lender.NewLender(s)
	q := NewQueuer(s, l)
	q.AddJobs()
}
