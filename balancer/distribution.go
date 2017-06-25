package balancer

import (
	"fmt"
)

var _ = fmt.Println

// AddUser will add a user to a bee and the BasePool
func (h *Hive) AddUser(u *User) error {
	// Find the Slave with the least on this exchange
	bees := h.Slaves.GetAndLockAllBees()
	defer h.Slaves.RUnlock()
	lowest := 0
	var candidate *Bee
	for _, b := range bees {
		exCount, _ := b.GetUnsafeExchangeCount(u.Exchange)
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
		// Add to phone book
		h.Slaves.AddUserUnsafe(u.Username, u.Exchange, candidate.ID)
	} else {
		h.BaseSlave.ChangeUser(u, true, true)
	}

	return nil
}

// RemoveUser will remove the user from any bees that have this user
// It will also remove from the basepool
func (h *Hive) RemoveUser(email string, exchange int) error {
	h.Slaves.Lock()
	defer h.Slaves.Unlock()

	u := User{Username: email, Exchange: exchange}
	p := NewChangeUserParcel(h.BaseSlave.ID, u, false, false)
	h.BaseSlave.SendChannel <- p

	beeID, ok := h.Slaves.GetUserUnsafe(email, exchange)
	if !ok {
		return fmt.Errorf("No bee found that has this user.")
	}
	p.ID = beeID

	ok = h.Slaves.SendParcelToUnsafe(beeID, p)
	if !ok {
		return fmt.Errorf("Send failed as the bee was not found")
	}

	return nil
}

func (h *Hive) MoveToBasePool(email string, exchange int) error {
	h.Slaves.Lock()
	defer h.Slaves.Unlock()

	u := User{Username: email, Exchange: exchange}
	pb := NewChangeUserParcel(h.BaseSlave.ID, u, true, true)

	beeID, ok := h.Slaves.GetUserUnsafe(email, exchange)
	if !ok {
		return fmt.Errorf("No bee found that has this user.")
	}
	p := NewChangeUserParcel(beeID, u, true, false)

	// Deactivate from bee
	h.Slaves.SendParcelToUnsafe(beeID, p)

	// Activate on BasePool
	baseSent := h.Slaves.SendParcelToUnsafe(h.BaseSlave.ID, pb)
	if !baseSent {
		return fmt.Errorf("Basepool slave could not be sent the message")
	}

	return nil
}
