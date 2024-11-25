package cloudian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
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

func NewClient(baseUrl string, tokenBase64 string) *Client {
	return &Client{
		baseURL:    baseUrl,
		httpClient: &http.Client{},
		token:      tokenBase64,
	}
}

// List all users of a group
func (client Client) ListUsers(ctx context.Context, groupId string, offsetUserId *string) ([]User, error) {
	var retVal []User

	limit := 100

	var offsetQueryParam = ""
	if offsetUserId != nil {
		offsetQueryParam = "&offset=" + *offsetUserId
	}

	url := client.baseURL + "/user/list?groupId=" + groupId + "&userType=all&userStatus=all&limit=" + strconv.Itoa(limit) + offsetQueryParam

	req, err := client.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("GET error creating list request: %w", err)
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET list users failed: %w", err)
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
		moreUsers, err := client.ListUsers(ctx, groupId, &users[limit].UserID)

		if err == nil {
			retVal = append(retVal, moreUsers...)
		}
	}

	return retVal, nil

}

// Delete a single user
func (client Client) DeleteUser(ctx context.Context, user User) error {
	url := client.baseURL + "/user?userId=" + user.UserID + "&groupId=" + user.GroupID + "&canonicalUserId=" + user.CanonicalUserID

	req, err := client.newRequest(ctx, url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close() // nolint:errcheck
	}

	return err
}

// Delete a group and all its members
func (client Client) DeleteGroupRecursive(ctx context.Context, groupId string) error {
	users, err := client.ListUsers(ctx, groupId, nil)
	if err != nil {
		return err
	}

	for _, user := range users {
		err := client.DeleteUser(ctx, user)
		if err != nil {
			return fmt.Errorf("Error deleting user: %w", err)
		}

	}

	return client.DeleteGroup(ctx, groupId)
}

// Deletes a group if it is without members
func (client Client) DeleteGroup(ctx context.Context, groupId string) error {
	url := client.baseURL + "/group?groupId=" + groupId

	req, err := client.newRequest(ctx, url, http.MethodDelete, nil)
	if err != nil {
		return err
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("DELETE to cloudian /group got: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	return nil
}

func (client Client) CreateGroup(ctx context.Context, group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	req, err := client.newRequest(ctx, url, http.MethodPost, &jsonData)
	if err != nil {
		return err
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("POST to cloudian /group: %w", err)
	}
	defer resp.Body.Close() // nolint:errcheck

	return err
}

func (client Client) UpdateGroup(ctx context.Context, group Group) error {

	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	// Create a context with a timeout
	req, err := client.newRequest(ctx, url, http.MethodPut, &jsonData)
	if err != nil {
		return err
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("PUT to cloudian /group: %w", err)
	}

	defer resp.Body.Close() // nolint:errcheck

	return nil
}

func (client Client) GetGroup(ctx context.Context, groupId string) (*Group, error) {
	url := client.baseURL + "/group?groupId=" + groupId

	req, err := client.newRequest(ctx, url, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("GET error: %w", err)
	}

	if resp != nil {
		defer resp.Body.Close() // nolint:errcheck
	}

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
	return nil, err
}

func (client Client) newRequest(ctx context.Context, url string, method string, body *[]byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(*body))
	if err != nil {
		return req, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

	return req, nil
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
