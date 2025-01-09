package cloudian

import (
	"encoding/json"
	"testing"
)

func TestSecretUnmarshal(t *testing.T) {
	jsonString := `[{"accessKey":"124","secretKey":"x+2","createDate":1735894172440,"active":true}]`

	var secrets []SecurityInfo
	err := json.Unmarshal([]byte(jsonString), &secrets)
	if err != nil {
		t.Errorf("Error deserializing from JSON: %v", err)
	}

	if secrets[0].SecretKey.String() != "********" {
		t.Errorf("Expected obfuscated string, got %v", secrets[0].SecretKey)
	}
}
