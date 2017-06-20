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

func NewRequestIDParcel(publicKey []byte) *Parcel {
	p := newParcel("", RequestIdentityParcel)
	p.Message = publicKey
	return p
}

type IDResponse struct {
	ID        string
	Users     []User
	PublicKey []byte
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
