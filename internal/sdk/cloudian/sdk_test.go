package cloudian

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func TestRealisticGroupSerialization(t *testing.T) {
	jsonString := `{
			"active": "true",
			"groupId": "QA",
			"groupName": "Quality Assurance Group",
			"ldapEnabled": false,
			"s3endpointshttp": ["ALL"],
			"s3endpointshttps": ["ALL"],
			"s3websiteendpoints": ["ALL"]
		}`

	var group groupInternal
	err := json.Unmarshal([]byte(jsonString), &group)
	if err != nil {
		t.Errorf("Error deserializing from JSON: %v", err)
	}

	if group.GroupIDInternal != "QA" {
		t.Errorf("Expected QA, got %v", group.GroupIDInternal)
	}
}

func TestUnmarshalUsers(t *testing.T) {
	jsonString := `[
		{
			"active": "true",
			"address1": "",
			"address2": "",
			"canonicalUserId": "fd221552ff4ddc857d7a9ca316bb8344",
			"city": "",
			"country": "",
			"emailAddr": "",
			"fullName": "Glory Bee",
			"groupId": "QA",
			"ldapEnabled": false,
			"phone": "",
			"state": "",
			"userId": "Glory",
			"userType": "User",
			"website": "",
			"zip": ""
		},
		{
			"active": "true",
			"address1": "",
			"address2": "",
			"canonicalUserId": "bd0796cd9746ef9cc4ef656ddaacfac4",
			"city": "",
			"country": "",
			"emailAddr": "",
			"fullName": "John Thompson",
			"groupId": "QA",
			"ldapEnabled": false,
			"phone": "",
			"state": "",
			"userId": "John",
			"userType": "User",
			"website": "",
			"zip": ""
			}]`

	var users []User
	err := json.Unmarshal([]byte(jsonString), &users)
	if err != nil {
		t.Errorf("Error deserializing users from JSON: %v", err)
	}

	if users[0].UserID != "Glory" {
		t.Errorf("Expected Glory as the userId of first user, got %v", users[0].UserID)
	}

	if users[1].UserID != "John" {
		t.Errorf("Expected John as the userId of second user, got %v", users[1].UserID)
	}

}

func (group groupInternal) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(groupInternal{
		ActiveInternal:             "true",
		GroupIDInternal:            randomString(16),
		GroupNameInternal:          randomString(32),
		LDAPEnabledInternal:        true,
		LDAPGroupInternal:          randomString(8),
		LDAPMatchAttributeInternal: randomString(8),
		LDAPSearchInternal:         randomString(8),
		LDAPSearchUserBaseInternal: randomString(8),
		LDAPServerURLInternal:      randomString(8),
		LDAPUserDNTemplateInternal: randomString(8),
		S3EndpointsHTTPInternal:    []string{randomString(8), randomString(8)},
		S3EndpointsHTTPSInternal:   []string{randomString(8), randomString(8)},
		S3WebSiteEndpointsInternal: []string{randomString(8), randomString(8)},
	})
}

func TestGenericError(t *testing.T) {
	err := errors.New("Random failure")

	if errors.Is(err, ErrNotFound) {
		t.Error("Expected not to be ErrNotFound")
	}
}

func TestWrappedErrNotFound(t *testing.T) {
	err := fmt.Errorf("wrap it: %w", ErrNotFound)

	if !errors.Is(err, ErrNotFound) {
		t.Error("Expected to be ErrNotFound")
	}
}

func TestGroupSerialization(t *testing.T) {
	f := func(group groupInternal) bool {
		data, err := json.Marshal(group)
		if err != nil {
			return false
		}

		var deserialized groupInternal
		if err = json.Unmarshal(data, &deserialized); err != nil {
			return false
		}

		return reflect.DeepEqual(group, deserialized)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

func randomString(length int) string {
	var sb strings.Builder
	runes := []rune(charset)
	for i := 0; i < length; i++ {
		sb.WriteRune(runes[rand.Intn(len(runes))])
	}
	return sb.String()
}
