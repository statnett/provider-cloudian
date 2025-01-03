package cloudian

type Secret string

func (s Secret) String() string {
	return "********"
}

// Gets the secret as a string.
func (s Secret) Reveal() string {
	return string(s)
}
