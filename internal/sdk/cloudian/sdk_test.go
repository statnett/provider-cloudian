package cloudian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

func TestGetGroup(t *testing.T) {
	expected := Group{
		GroupID: "QA",
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"groupId": "QA"}`))
	}))
	defer mockServer.Close()

	mockClient := &http.Client{}

	_, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, mockServer.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	cloudianClient := Client{
		baseURL:    mockServer.URL,
		httpClient: mockClient,
		authHeader: "",
	}

	group, err := cloudianClient.GetGroup(context.TODO(), "QA")

	if err != nil {
		t.Errorf("Error getting group: %v", err)
	}

	if diff := cmp.Diff(*group, expected); diff != "" {
		t.Errorf("GetGroup() mismatch (-want +got):\n%s", diff)
	}

}

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
		Active:             "true",
		GroupID:            randomString(16),
		GroupName:          randomString(32),
		LDAPEnabled:        true,
		LDAPGroup:          randomString(8),
		LDAPMatchAttribute: randomString(8),
		LDAPSearch:         randomString(8),
		LDAPSearchUserBase: randomString(8),
		LDAPServerURL:      randomString(8),
		LDAPUserDNTemplate: randomString(8),
		S3EndpointsHTTP:    []string{randomString(8), randomString(8)},
		S3EndpointsHTTPS:   []string{randomString(8), randomString(8)},
		S3WebSiteEndpoints: []string{randomString(8), randomString(8)},
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

func TestGroupToInternalRoundtrip(t *testing.T) {
	f := func(group groupInternal) bool {
		// Override fuzzed s3-endpoint entries with known values
		group.S3EndpointsHTTP = []string{"ALL"}
		group.S3EndpointsHTTPS = []string{"ALL"}
		group.S3WebSiteEndpoints = []string{"ALL"}

		roundtrip := toInternal(fromInternal(group))
		return reflect.DeepEqual(group, roundtrip)
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
