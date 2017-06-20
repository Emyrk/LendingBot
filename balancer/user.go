package balancer

import (
	"time"
)

type User struct {
	SlaveID   string
	LastTouch time.Time
	Username  string
	Active    bool
}
