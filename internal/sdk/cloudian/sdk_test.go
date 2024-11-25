package cloudian

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-. "

func randomString(length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func TestRealisticGroupSerialization(t *testing.T) {
	jsonString := `{
			"active": "true",
			"groupId": "QA",
			"groupName": "Quality Assurance Group",
			"ldapEnabled": false,
			"ldapGroup": "",
			"ldapMatchAttribute": "",
			"ldapSearch": "",
			"ldapSearchUserBase": "",
			"ldapServerURL": "",
			"ldapUserDNTemplate": "",
			"s3endpointshttp": ["ALL"],
			"s3endpointshttps": ["ALL"],
			"s3websiteendpoints": ["ALL"]
		}`

	group, err := unmarshalGroupJson([]byte(jsonString))
	if err != nil {
		t.Errorf("Error deserializing from JSON: %v", err)
	}

	if group.GroupID != "QA" {
		t.Errorf("Expected QA, got %v", group.GroupID)
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

	users, err := unmarshalUsersJson([]byte(jsonString))
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

func (group Group) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Group{
		Active:             "true",
		GroupID:            randomString(16),
		GroupName:          randomString(32),
		LDAPEnabled:        false,
		LDAPGroup:          randomString(8),
		LDAPMatchAttribute: "",
		LDAPSearch:         "",
		LDAPSearchUserBase: "",
		LDAPServerURL:      "",
		LDAPUserDNTemplate: "",
		S3endpointsHTTP:    []string{"ALL"},
		S3endpointsHTTPS:   []string{"ALL"},
		S3websiteendpoints: []string{"ALL"},
	})
}

func TestGroupSerialization(t *testing.T) {
	f := func(group Group) bool {
		data, err := marshalGroup(group)
		if err != nil {
			return false
		}

		deserialized, err := unmarshalGroupJson(data)
		if err != nil {
			return false
		}

		return reflect.DeepEqual(group, deserialized)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
