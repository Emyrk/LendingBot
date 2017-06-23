package balancer

import (
	"fmt"
)

var _ = fmt.Println

func (h *Hive) AddUser(u *User) error {
	// Find the Slave with the least on this exchange
	bees := h.Slaves.GetAndLockAllBees()
	lowest := 0
	var candidate *Bee
	for _, b := range bees {
		exCount, _ := b.GetExchangeCount(u.Exchange)
		fmt.Println(exCount)
		if candidate == nil {
			candidate = b
			lowest = exCount
		} else {
			if exCount < lowest {
				candidate = b
			}
		}
	}

	// Have a candidate
	if candidate != nil {
		candidate.ChangeUser(u, true, true)
		h.BaseSlave.ChangeUser(u, true, false)
	} else {
		h.BaseSlave.ChangeUser(u, true, true)
	}

	return nil
}
