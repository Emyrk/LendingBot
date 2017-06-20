package balancer

const (
	RequestIdentity int = iota
	ResponseIdentity
)

type Parcel struct {
	// Header
	ID   string
	Type int

	// Body
	Message interface{}
}

func NewRequestIDParcel() *Parcel {
	return newParcel("", RequestIdentity)
}

type IDResponse struct {
	ID    string
	Users []User
}

func NewResponseIDParcel(id string, users []User) *Parcel {
	p := newParcel(id, ResponseIdentity)

	m := new(IDResponse)
	m.ID = id
	m.Users = users
	p.Message = m

	return
}

func newParcel(id string, t int) *Parcel {
	p := new(Parcel)
	p.Type = t
	p.ID = id

	return p
}
