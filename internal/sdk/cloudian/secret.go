package cloudian

type Secret string

func (s Secret) String() string {
	return "********"
}
