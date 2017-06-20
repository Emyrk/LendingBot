package balancer

import (
	"sort"
	"strings"
	"time"
)

type UserList []User

func (slice UserList) Len() int {
	return len(slice)
}

func (slice UserList) Less(i, j int) bool {
	if strings.Compare(slice[i].Username, slice[j].Username) < 0 {
		return true
	}
	return false
}

func (slice UserList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// CompareUserList returns true if the same
func CompareUserList(a []User, b []User) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Sort(UserList(a))
	sort.Sort(UserList(b))

	for i, _ := range a {
		if !a[i].IsSimilar(&b[i]) {
			return false
		}
	}
	return true
}

type User struct {
	SlaveID   string
	LastTouch time.Time
	Username  string
	Active    bool
}

func (a *User) IsSimilar(b *User) bool {
	if a.SlaveID != b.SlaveID {
		return false
	}

	if a.Username != b.Username {
		return false
	}

	if a.Active != b.Active {
		return false
	}
	return true
}
