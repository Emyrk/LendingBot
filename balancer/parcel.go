package balancer

import (
	"encoding/json"
	"time"
)

const (
	RequestIdentityParcel int = iota
	ResponseIdentityParcel

	AssignmentParcel

	HeartbeatParcel

	ChangeUserParcel
)

type Parcel struct {
	// Header
	ID   string // Serves as 'To:' and 'From:' ID. It always refers to a Bee
	Type int

	// Body
	Message []byte
}

func NewRequestIDParcel(publicKey []byte) *Parcel {
	p := newParcel("", RequestIdentityParcel)
	p.Message = publicKey
	return p
}

type IDResponse struct {
	ID        string
	Users     []*User
	PublicKey []byte
}

func NewResponseIDParcel(id string, users []*User, public []byte) *Parcel {
	p := newParcel(id, ResponseIdentityParcel)

	m := new(IDResponse)
	m.ID = id
	m.Users = users
	m.PublicKey = public

	msg, _ := json.Marshal(m)
	p.Message = msg

	return p
}

type Assignment struct {
	Users []*User
}

func NewAssignment(id string, a Assignment) *Parcel {
	p := newParcel(id, AssignmentParcel)
	msg, _ := json.Marshal(a)
	p.Message = msg

	return p
}

type NewChangeUser struct {
	U      User
	Add    bool
	Active bool
}

func NewChangeUserParcel(id string, u User, add, active bool) *Parcel {
	p := newParcel(id, ChangeUserParcel)
	m := new(NewChangeUser)
	m.U = u
	m.Active = active
	m.Add = add

	msg, _ := json.Marshal(m)
	p.Message = msg

	return p
}

type Heartbeat struct {
	SentTime    time.Time
	Users       []*User
	ApiRate     float64
	LoanJobRate float64
}

func NewHeartbeat(id string, h Heartbeat) *Parcel {
	p := newParcel(id, HeartbeatParcel)
	msg, _ := json.Marshal(&h)
	p.Message = msg

	return p
}

func newParcel(id string, t int) *Parcel {
	p := new(Parcel)
	p.Type = t
	p.ID = id

	return p
}
