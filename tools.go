package forza

type Tools struct {
	name    string
	execute interface{}
}

func NewTools() *Tools {
	return &Tools{}
}
