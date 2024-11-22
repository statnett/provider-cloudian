package cloudian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

func NewClient(baseUrl string, tokenBase64 string) *Client {
	return &Client{
		baseURL:    baseUrl,
		httpClient: &http.Client{},
		token:      tokenBase64,
	}
}

func (client Client) headerModifier(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)
}

type Group struct {
	Active             string   `json:"active"`
	GroupID            string   `json:"groupId"`
	GroupName          string   `json:"groupName"`
	LDAPEnabled        bool     `json:"ldapEnabled"`
	LDAPGroup          string   `json:"ldapGroup"`
	LDAPMatchAttribute string   `json:"ldapMatchAttribute"`
	LDAPSearch         string   `json:"ldapSearch"`
	LDAPSearchUserBase string   `json:"ldapSearchUserBase"`
	LDAPServerURL      string   `json:"ldapServerURL"`
	LDAPUserDNTemplate string   `json:"ldapUserDNTemplate"`
	S3endpointshttp    []string `json:"s3endpointshttp"`
	S3endpointshttps   []string `json:"s3endpointshttps"`
	S3websiteendpoints []string `json:"s3websiteendpoints"`
}

type User struct {
	UserID          string `json:"userId"`
	GroupID         string `json:"groupId"`
	CanonicalUserID string `json:"canonicalUserId"`
}

func marshalGroup(group Group) ([]byte, error) {
	return json.Marshal(group)
}

func unmarshalGroupJson(data []byte) (Group, error) {
	var group Group
	err := json.Unmarshal(data, &group)
	return group, err
}

func unmarshalUsersJson(data []byte) ([]User, error) {
	var users []User
	err := json.Unmarshal(data, &users)
	return users, err
}

// List all users of a group
func (client Client) ListUsers(groupId string, offsetUserId *string) ([]User, error) {
	var retVal []User

	limit := 100

	var offsetQueryParam = ""
	if offsetUserId != nil {
		offsetQueryParam = "&offset=" + *offsetUserId
	}

	url := client.baseURL + "/user/list?groupId=" + groupId + "&userType=all&userStatus=all&limit=" + strconv.Itoa(limit) + offsetQueryParam

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GET error creating list request: %w", err)
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("GET list users failed: %w", err)
	}
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close() // nolint:errcheck
		if err != nil {
			return nil, fmt.Errorf("GET reading list users response body failed: %w", err)
		}

		users, err := unmarshalUsersJson(body)
		if err != nil {
			return nil, fmt.Errorf("GET unmarshal users response body failed: %w", err)
		}

		retVal = append(retVal, users...)

		// list users is a paginated API endpoint, so we need to check the limit and use an offset to fetch more
		if len(users) > limit {
			// There is some ambiguity in the GET /user/list endpoint documentation, but it seems
			// that UserId is the correct key for this parameter (and not CanonicalUserId)
			// Fetch more results
			moreUsers, err := client.ListUsers(groupId, &users[limit].UserID)

			if err == nil {
				retVal = append(retVal, moreUsers...)
			}
		}

		return retVal, nil
	} else {
		return nil, fmt.Errorf("GET list users failed: %w", err)
	}

}

// Delete a single user
func (client Client) DeleteUser(user User) error {
	url := client.baseURL + "/user?userId=" + user.UserID + "&groupId=" + user.GroupID + "&canonicalUserId=" + user.CanonicalUserID

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("DELETE error creating request: %w", err)
	}

	client.headerModifier(req)

	resp, err := client.httpClient.Do(req)

	if resp != nil && err == nil {
		// Cloudian does not return a payload for this DELETE, but we can echo it to the callsite if all went well
		defer resp.Body.Close() // nolint:errcheck

		return nil
	}
	return err
}

// Delete a group and all its members
func (client Client) DeleteGroupRecursive(groupId string) error {
	users, err := client.ListUsers(groupId, nil)
	if err != nil {
		return err
	}

	for _, user := range users {
			err := client.DeleteUser(user)
			if err != nil {
				return fmt.Errorf("Error deleting user: %w", err)
			}
		}

		return client.DeleteGroup(groupId)
	}

	return err
}

// Deletes a group if it is without members
func (client Client) DeleteGroup(groupId string) error {
	url := client.baseURL + "/group?groupId=" + groupId

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("DELETE error creating request: %w", err)
	}

	client.headerModifier(req)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.Itoa(resp.StatusCode)
		return fmt.Errorf("DELETE to cloudian /group got status code [%s]: %w", statusErrStr, err)
	}
	defer resp.Body.Close() // nolint:errcheck

	return nil
}

func (client Client) CreateGroup(group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST error creating request: %w", err)
	}

	client.headerModifier(req)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
		return fmt.Errorf("POST to cloudian /group got status code [%s]: %w", statusErrStr, err)
	}
	defer resp.Body.Close() // nolint:errcheck

	return err
}

func (client Client) UpdateGroup(group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("PUT error creating request: %w", err)
	}

	client.headerModifier(req)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
		return fmt.Errorf("PUT to cloudian /group got status code [%s]: %w", statusErrStr, err)
	}

	defer resp.Body.Close() // nolint:errcheck

	return nil
}

func (client Client) GetGroup(groupId string) (*Group, error) {
	url := client.baseURL + "/group?groupId=" + groupId

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GET error creating request: %w", err)
	}

	client.headerModifier(req)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET error: %w", err)
	}

	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("GET reading response body failed: %w", err)
		}

		group, err := unmarshalGroupJson(body)
		if err != nil {
			return nil, fmt.Errorf("GET unmarshal response body failed: %w", err)
		}

		return &group, nil
	}

	// Cloudian-API returns 204 if the group does not exist
	return nil, nil
}
