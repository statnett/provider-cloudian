package cloudian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	var expected []User
	for i := 0; i < 500; i++ {
		expected = append(expected, User{GroupID: "QA", UserID: strconv.Itoa(i)})
	}

	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		index := 0

		if offset := r.URL.Query().Get("offset"); offset != "" {
			var err error
			index, err = strconv.Atoi(r.URL.Query().Get("offset"))
			if err != nil {
				panic(err)
			}
		}

		// return one more than limit to indicate more pages
		end := index + ListLimit + 1
		if end > len(expected) {
			end = len(expected)
		}
		json.NewEncoder(w).Encode(expected[index:end])
	})
	defer testServer.Close()

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
