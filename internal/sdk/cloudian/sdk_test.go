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
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(groupInternal{GroupID: "QA", Active: "true"})
	})
	defer testServer.Close()

	expected := Group{
		GroupID: "QA",
		Active:  true,
	}

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
	expected := &SecurityInfo{AccessKey: "123", SecretKey: "abc"}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(expected)
	})
	defer testServer.Close()

	credentials, err := cloudianClient.GetUserCredentials(context.TODO(), "123")

	if err != nil {
		t.Errorf("Error getting credentials: %v", err)
	}

	if diff := cmp.Diff(expected, credentials); diff != "" {
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
	expected := []User{{GroupID: "QA", UserID: "user1"}}
	cloudianClient, testServer := mockBy(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(expected)
	})
	defer testServer.Close()

	users, err := cloudianClient.ListUsers(context.TODO(), "QA", nil)

	if err != nil {
		t.Errorf("Error listing users: %v", err)
	}

	if diff := cmp.Diff(expected, users); diff != "" {
		t.Errorf("ListUsers() mismatch (-want +got):\n%s", diff)
	}
}

func mockBy(handler http.HandlerFunc) (*Client, *httptest.Server) {
	mockServer := httptest.NewServer(handler)

	return NewClient(mockServer.URL, ""), mockServer
}
