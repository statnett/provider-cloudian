package cloudian

import (
	"encoding/json"
)

type Secret struct {
	value string
}

func (s *Secret) String() string {
	return "********"
}

// Gets the secret as a string.
func (s *Secret) Reveal() string {
	if s != nil {
		return s.value
	}
	return ""
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = Secret{str}
	return nil
}
