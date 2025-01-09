package cloudian

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	authHeader string
}

type Group struct {
	Active             bool   `json:"active"`
	GroupID            string `json:"groupId"`
	GroupName          string `json:"groupName"`
	LDAPEnabled        bool   `json:"ldapEnabled"`
	LDAPGroup          string `json:"ldapGroup"`
	LDAPMatchAttribute string `json:"ldapMatchAttribute"`
	LDAPSearch         string `json:"ldapSearch"`
	LDAPSearchUserBase string `json:"ldapSearchUserBase"`
	LDAPServerURL      string `json:"ldapServerURL"`
	LDAPUserDNTemplate string `json:"ldapUserDNTemplate"`
}

// groupInternal is the SDK's internal representation of a cloudion group.
// Fields must be exported (uppercase) to allow json marshalling.
type groupInternal struct {
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
	S3EndpointsHTTP    []string `json:"s3endpointshttp"`
	S3EndpointsHTTPS   []string `json:"s3endpointshttps"`
	S3WebSiteEndpoints []string `json:"s3websiteendpoints"`
}

// NewGroup creates an empty cloudian group with the given ID.
func NewGroup(groupID string) Group {
	return Group{
		GroupID: groupID,
	}
}

func toInternal(g Group) groupInternal {
	return groupInternal{
		Active:             strconv.FormatBool(g.Active),
		GroupID:            g.GroupID,
		GroupName:          g.GroupName,
		LDAPEnabled:        g.LDAPEnabled,
		LDAPGroup:          g.LDAPGroup,
		LDAPMatchAttribute: g.LDAPMatchAttribute,
		LDAPSearch:         g.LDAPSearch,
		LDAPSearchUserBase: g.LDAPSearchUserBase,
		LDAPServerURL:      g.LDAPServerURL,
		LDAPUserDNTemplate: g.LDAPUserDNTemplate,
		S3EndpointsHTTP:    []string{"ALL"},
		S3EndpointsHTTPS:   []string{"ALL"},
		S3WebSiteEndpoints: []string{"ALL"},
	}
}

func fromInternal(g groupInternal) Group {
	return Group{
		Active:             g.Active == "true",
		GroupID:            g.GroupID,
		GroupName:          g.GroupName,
		LDAPEnabled:        g.LDAPEnabled,
		LDAPGroup:          g.LDAPGroup,
		LDAPMatchAttribute: g.LDAPMatchAttribute,
		LDAPSearch:         g.LDAPSearch,
		LDAPSearchUserBase: g.LDAPSearchUserBase,
		LDAPServerURL:      g.LDAPServerURL,
		LDAPUserDNTemplate: g.LDAPUserDNTemplate,
	}
}

type User struct {
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
}

type userInternal struct {
	UserID   string `json:"userId"`
	GroupID  string `json:"groupId"`
	UserType string `json:"userType"`
}

func toInternalUser(u User) userInternal {
	return userInternal{
		UserID:   u.UserID,
		GroupID:  u.GroupID,
		UserType: "User",
	}
}

// SecurityInfo is the Cloudian API's term for secure credentials
type SecurityInfo struct {
	AccessKey Secret `json:"accessKey"`
	SecretKey Secret `json:"secretKey"`
}

var ErrNotFound = errors.New("not found")

// WithInsecureTLSVerify skips the TLS validation of the server certificate when `insecure` is true.
func WithInsecureTLSVerify(insecure bool) func(*Client) {
	return func(c *Client) {
		c.httpClient = &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, // nolint:gosec
		}}
	}
}

func NewClient(baseURL string, authHeader string, opts ...func(*Client)) *Client {
	c := &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		authHeader: authHeader,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// List all users of a group.
func (client Client) ListUsers(ctx context.Context, groupId string, offsetUserId *string) ([]User, error) {
	var retVal []User

	limit := 100
	params := map[string]string{
		"groupId":    groupId,
		"userType":   "all",
		"userStatus": "all",
		"limit":      strconv.Itoa(limit),
	}
	if offsetUserId != nil {
		params["offset"] = *offsetUserId
	}

	resp, err := client.doRequest(ctx, http.MethodGet, "/user/list", params, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GET reading list users response body failed: %w", err)
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("GET unmarshal users response body failed: %w", err)
	}

	retVal = append(retVal, users...)

	// list users is a paginated API endpoint, so we need to check the limit and use an offset to fetch more
	if len(users) > limit {
		retVal = retVal[0 : len(retVal)-1] // Remove the last element, which is the offset
		// There is some ambiguity in the GET /user/list endpoint documentation, but it seems
		// that UserId is the correct key for this parameter
		// Fetch more results
		moreUsers, err := client.ListUsers(ctx, groupId, &users[limit].UserID)
		if err != nil {
			return nil, fmt.Errorf("GET list users failed: %w", err)
		}

		retVal = append(retVal, moreUsers...)
	}

	return retVal, nil

}

// Delete a single user. Errors if the user does not exist.
func (client Client) DeleteUser(ctx context.Context, user User) error {
	resp, err := client.doRequest(ctx, http.MethodDelete, "/user",
		map[string]string{"groupId": user.GroupID, "userId": user.UserID}, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE unexpected status. Failure: %d", resp.StatusCode)
	}

}

// Create a single user of type `User` into a groupId
func (client Client) CreateUser(ctx context.Context, user User) error {
	jsonData, err := json.Marshal(toInternalUser(user))
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	resp, err := client.doRequest(ctx, http.MethodPut, "/user", nil, jsonData)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// CreateUserCredentials creates a new set of credentials for a user.
func (client Client) CreateUserCredentials(ctx context.Context, user User) (*SecurityInfo, error) {
	resp, err := client.doRequest(ctx, http.MethodPut, "/user/credentials",
		map[string]string{"groupId": user.GroupID, "userId": user.UserID}, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading create credentials response: %w", err)
		}

		var securityInfo SecurityInfo
		if err := json.Unmarshal(body, &securityInfo); err != nil {
			return nil, fmt.Errorf("error parsing create credentials response: %w", err)
		}

		return &securityInfo, nil
	default:
		return nil, fmt.Errorf("error: create credentials unexpected status code: %d", resp.StatusCode)
	}
}

// GetUserCredentials fetches a set of credentials for a user.
func (client Client) GetUserCredentials(ctx context.Context, accessKey string) (*SecurityInfo, error) {
	resp, err := client.doRequest(ctx, http.MethodGet, "/user/credentials",
		map[string]string{"accessKey": accessKey}, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading get credentials response: %w", err)
		}

		var securityInfo SecurityInfo
		if err := json.Unmarshal(body, &securityInfo); err != nil {
			return nil, fmt.Errorf("error parsing get credentials response: %w", err)
		}

		return &securityInfo, nil
	case 204:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error: get credentials unexpected status code: %d", resp.StatusCode)
	}
}

// ListUserCredentials fetches all the credentials of a user.
func (client Client) ListUserCredentials(ctx context.Context, user User) ([]SecurityInfo, error) {
	resp, err := client.doRequest(ctx, http.MethodGet, "/user/credentials/list",
		map[string]string{"groupId": user.GroupID, "userId": user.UserID}, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading list credentials response: %w", err)
		}

		var securityInfo []SecurityInfo
		if err := json.Unmarshal(body, &securityInfo); err != nil {
			return nil, fmt.Errorf("error parsing list credentials response: %w", err)
		}

		return securityInfo, nil
	case 204:
		// Cloudian-API returns 204 if no security credentials found
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error: list credentials unexpected status code: %d", resp.StatusCode)
	}
}

// DeleteUserCredentials deletes a set of credentials for a user.
func (client Client) DeleteUserCredentials(ctx context.Context, accessKey string) error {
	resp, err := client.doRequest(ctx, http.MethodDelete, "/user/credentials",
		map[string]string{"accessKey": accessKey}, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE unexpected status. Failure: %d", resp.StatusCode)
	}
}

// Delete a group and all its members.
func (client Client) DeleteGroupRecursive(ctx context.Context, groupId string) error {
	users, err := client.ListUsers(ctx, groupId, nil)
	if err != nil {
		return fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range users {
		if err := client.DeleteUser(ctx, user); err != nil {
			return fmt.Errorf("error deleting user: %w", err)
		}
	}

	return client.DeleteGroup(ctx, groupId)
}

// Deletes a group if it is without members.
func (client Client) DeleteGroup(ctx context.Context, groupId string) error {
	resp, err := client.doRequest(ctx, http.MethodDelete, "/group",
		map[string]string{"groupId": groupId}, nil)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// Creates a group.
func (client Client) CreateGroup(ctx context.Context, group Group) error {
	jsonData, err := json.Marshal(toInternal(group))
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	resp, err := client.doRequest(ctx, http.MethodPut, "/group", nil, jsonData)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// Updates a group if it does not exists.
func (client Client) UpdateGroup(ctx context.Context, group Group) error {
	jsonData, err := json.Marshal(toInternal(group))
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a context with a timeout
	resp, err := client.doRequest(ctx, http.MethodPost, "/group", nil, jsonData)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

// Get a group. Returns an error even in the case of a group not found.
// This error can then be checked against ErrNotFound: errors.Is(err, ErrNotFound)
func (client Client) GetGroup(ctx context.Context, groupId string) (*Group, error) {
	resp, err := client.doRequest(ctx, http.MethodGet, "/group",
		map[string]string{"groupId": groupId}, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	switch resp.StatusCode {
	case 200:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("GET reading response body failed: %w", err)
		}

		var group groupInternal
		if err := json.Unmarshal(body, &group); err != nil {
			return nil, fmt.Errorf("GET unmarshal response body failed: %w", err)
		}

		retVal := fromInternal(group)
		return &retVal, nil
	case 204:
		// Cloudian-API returns 204 if the group does not exist
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("GET unexpected status. Failure: %w", err)
	}
}

func (client Client) doRequest(ctx context.Context, method string, url string, query map[string]string, body []byte) (*http.Response, error) {
	var buffer io.Reader = nil
	if body != nil {
		buffer = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, client.baseURL+url, buffer)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", client.authHeader)

	q := req.URL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return client.httpClient.Do(req)
}
