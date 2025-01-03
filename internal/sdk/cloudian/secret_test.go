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

	if string(secrets[0].AccessKey) != "124" {
		t.Errorf("Expected string equality to 124, got %v", secrets[0].AccessKey)
	}

	if secrets[0].AccessKey.String() != "********" {
		t.Errorf("Expected obfuscated string, got %v", secrets[0].SecretKey)
	}
}
