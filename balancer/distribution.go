package balancer

import (
	"fmt"
)

var _ = fmt.Println

// AddUser will add a user to a bee and the BasePool
func (h *Hive) AddUser(u *User) error {
	var err error

	// Ensure API key exists
	if u.AccessKey == "" {
		u, err = h.parent.IRS.GetFullUser(u.Username, u.Exchange)
		if err != nil {
			return err
		}
		if u == nil {
			return fmt.Errorf("User not found in db")
		}
	}

	// Find the Slave with the least on this exchange
	bees := h.Slaves.GetAndLockAllBees(false)
	defer h.Slaves.Unlock()
	lowest := 0
	var candidate *Bee
	for _, b := range bees {
		if b.Status != Online {
			continue
		}
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
		candidate.ChangeUserUnsafe(u, true, true)
		// Add to phone book
		h.Slaves.AddUserUnsafe(u.Username, u.Exchange, candidate.ID)
	} else {
		return fmt.Errorf("No slaves that the user can be added too")
	}

	return nil
}

// RemoveUser will remove the user from any bees that have this user
// It will also remove from the basepool
func (h *Hive) RemoveUser(email string, exchange int) error {
	h.Slaves.Lock()
	defer h.Slaves.Unlock()

	u := User{Username: email, Exchange: exchange}
	p := NewChangeUserParcel("", u, false, false)

	beeID, ok := h.Slaves.GetUserUnsafe(email, exchange)
	if !ok {
		return fmt.Errorf("No bee found that has this user.")
	}
	p.ID = beeID

	ok = h.Slaves.RemoveUserUnsafe(email, exchange)
	if !ok {
		// User was not found to be deleted
	}

	ok = h.Slaves.SendParcelToUnsafe(beeID, p)
	if !ok {
		return fmt.Errorf("Send failed as the bee was not found")
	}

	return nil
}

func (h *Hive) RemoveUserFromBee(id string, email string, exchange int) error {
	h.Slaves.Lock()
	defer h.Slaves.Unlock()

	u := User{Username: email, Exchange: exchange}
	p := NewChangeUserParcel(id, u, false, false)
	ok := h.Slaves.SendParcelToUnsafe(id, p)
	if !ok {
		return fmt.Errorf("Send failed as the bee was not found")
	}

	return nil
}

// func (h *Hive) MoveToBasePool(email string, exchange int) error {
// 	h.Slaves.Lock()
// 	defer h.Slaves.Unlock()

// 	u := User{Username: email, Exchange: exchange}
// 	pb := NewChangeUserParcel(h.BaseSlave.ID, u, true, true)

// 	beeID, ok := h.Slaves.GetUserUnsafe(email, exchange)
// 	if !ok {
// 		return fmt.Errorf("No bee found that has this user.")
// 	}
// 	p := NewChangeUserParcel(beeID, u, true, false)

// 	// Deactivate from bee
// 	h.Slaves.SendParcelToUnsafe(beeID, p)

// 	// Activate on BasePool
// 	baseSent := h.Slaves.SendParcelToUnsafe(h.BaseSlave.ID, pb)
// 	if !baseSent {
// 		return fmt.Errorf("Basepool slave could not be sent the message")
// 	}

// 	return nil
// }
