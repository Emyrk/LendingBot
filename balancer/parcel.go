package balancer

const (
	RequestIdentityParcel int = iota
	ResponseIdentityParcel

	AssignmentParcel
)

type Parcel struct {
	// Header
	ID   string
	Type int

	// Body
	Message interface{}
}

func NewRequestIDParcel() *Parcel {
	return newParcel("", RequestIdentityParcel)
}

type IDResponse struct {
	ID    string
	Users []User
}

func NewResponseIDParcel(id string, users []User) *Parcel {
	p := newParcel(id, ResponseIdentityParcel)

	m := new(IDResponse)
	m.ID = id
	m.Users = users
	p.Message = m

	return p
}

type Assignment struct {
	Users []User
}

func NewAssignment(id string, a Assignment) *Parcel {
	p := newParcel(id, AssignmentParcel)
	p.Message = a

	return p
}

func newParcel(id string, t int) *Parcel {
	p := new(Parcel)
	p.Type = t
	p.ID = id

	return p
}
