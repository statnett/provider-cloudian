package cloudian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestGetGroup(t *testing.T) {
	expected := Group{
		GroupID: "QA",
		Active:  true,
	}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toInternal(expected))
	})
	defer testServer.Close()

	group, err := cloudianClient.GetGroup(context.TODO(), "QA")
	if err != nil {
		t.Errorf("Error getting group: %v", err)
	}
	if diff := cmp.Diff(expected, *group); diff != "" {
		t.Errorf("GetGroup() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetGroupNotFound(t *testing.T) {
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer testServer.Close()

	_, err := cloudianClient.GetGroup(context.TODO(), "QA")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected error to be ErrNotFound")
	}
}

func TestCreateCredentials(t *testing.T) {
	expected := SecurityInfo{AccessKey: "123", SecretKey: "abc"}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(expected)
	})
	defer testServer.Close()

	credentials, err := cloudianClient.CreateUserCredentials(context.TODO(), User{GroupID: "QA", UserID: "user1"})
	if err != nil {
		t.Errorf("Error creating credentials: %v", err)
	}
	if diff := cmp.Diff(expected, *credentials); diff != "" {
		t.Errorf("CreateUserCredentials() mismatch (-want +got):\n%s", diff)
	}
}

func TestGetUserCredentials(t *testing.T) {
	expected := SecurityInfo{AccessKey: "123", SecretKey: "abc"}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(expected)
	})
	defer testServer.Close()

	credentials, err := cloudianClient.GetUserCredentials(context.TODO(), "123")
	if err != nil {
		t.Errorf("Error getting credentials: %v", err)
	}
	if diff := cmp.Diff(expected, *credentials); diff != "" {
		t.Errorf("GetUserCredentials() mismatch (-want +got):\n%s", diff)
	}
}

func TestListUserCredentials(t *testing.T) {
	expected := []SecurityInfo{
		{AccessKey: "123", SecretKey: "abc"},
		{AccessKey: "456", SecretKey: "def"},
	}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(expected)
	})
	defer testServer.Close()

	credentials, err := cloudianClient.ListUserCredentials(
		context.TODO(), User{UserID: "", GroupID: ""},
	)
	if err != nil {
		t.Errorf("Error listing credentials: %v", err)
	}
	if diff := cmp.Diff(expected, credentials); diff != "" {
		t.Errorf("ListUserCredentials() mismatch (-want +got):\n%s", diff)
	}
}

func TestListUsers(t *testing.T) {
	mkUsers := func(offset, n int) []User {
		users := make([]User, 0)
		for i := offset; i < n; i++ {
			users = append(users, User{GroupID: "QA", UserID: fmt.Sprintf("user%d", i)})
		}
		return users
	}
	// We pretend 102 users exist in the cloudian server
	expected := mkUsers(0, 102)

	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "" {
			// return 101 users in the first batch (the 101th indicating there are "more results")
			json.NewEncoder(w).Encode(expected[:101])
		} else {
			// return the two last users (0-indexed) as the last batch
			json.NewEncoder(w).Encode(expected[100:])
		}
	})
	defer testServer.Close()

	// the first 101 users (indicating "more results" from server)
	users, err := cloudianClient.ListUsers(context.Background(), "QA", nil)
	if err != nil {
		t.Errorf("Error listing users: %v", err)
	}
	if diff := cmp.Diff(expected, users); diff != "" {
		t.Errorf("ListUsers() mismatch without offset (-want +got):\n%s", diff)
	}

}

func mockBy(handler http.HandlerFunc) (*Client, *httptest.Server) {
	mockServer := httptest.NewServer(handler)
	return NewClient(mockServer.URL, ""), mockServer
}
