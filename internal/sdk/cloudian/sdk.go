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

// fields must be exported (uppercase) to allow json marshalling
type groupInternal struct {
	ActiveInternal             string   `json:"active"`
	GroupIDInternal            string   `json:"groupId"`
	GroupNameInternal          string   `json:"groupName"`
	LDAPEnabledInternal        bool     `json:"ldapEnabled"`
	LDAPGroupInternal          string   `json:"ldapGroup"`
	LDAPMatchAttributeInternal string   `json:"ldapMatchAttribute"`
	LDAPSearchInternal         string   `json:"ldapSearch"`
	LDAPSearchUserBaseInternal string   `json:"ldapSearchUserBase"`
	LDAPServerURLInternal      string   `json:"ldapServerURL"`
	LDAPUserDNTemplateInternal string   `json:"ldapUserDNTemplate"`
	S3EndpointsHTTPInternal    []string `json:"s3endpointshttp"`
	S3EndpointsHTTPSInternal   []string `json:"s3endpointshttps"`
	S3WebSiteEndpointsInternal []string `json:"s3websiteendpoints"`
}

func toInternal(g Group) groupInternal {
	return groupInternal{
		ActiveInternal:             strconv.FormatBool(g.Active),
		GroupIDInternal:            g.GroupID,
		GroupNameInternal:          g.GroupName,
		LDAPEnabledInternal:        g.LDAPEnabled,
		LDAPGroupInternal:          g.LDAPGroup,
		LDAPMatchAttributeInternal: g.LDAPMatchAttribute,
		LDAPSearchInternal:         g.LDAPSearch,
		LDAPSearchUserBaseInternal: g.LDAPSearchUserBase,
		LDAPServerURLInternal:      g.LDAPServerURL,
		LDAPUserDNTemplateInternal: g.LDAPUserDNTemplate,
		S3EndpointsHTTPInternal:    []string{"ALL"},
		S3EndpointsHTTPSInternal:   []string{"ALL"},
		S3WebSiteEndpointsInternal: []string{"ALL"},
	}
}

func fromInternal(g groupInternal) Group {
	return Group{
		Active:             g.ActiveInternal == "true",
		GroupID:            g.GroupIDInternal,
		GroupName:          g.GroupNameInternal,
		LDAPEnabled:        g.LDAPEnabledInternal,
		LDAPGroup:          g.LDAPGroupInternal,
		LDAPMatchAttribute: g.LDAPMatchAttributeInternal,
		LDAPSearch:         g.LDAPSearchInternal,
		LDAPSearchUserBase: g.LDAPSearchUserBaseInternal,
		LDAPServerURL:      g.LDAPServerURLInternal,
		LDAPUserDNTemplate: g.LDAPUserDNTemplateInternal,
	}
}

type User struct {
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
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

	offsetQueryParam := ""
	if offsetUserId != nil {
		offsetQueryParam = "&offset=" + *offsetUserId
	}

	url := client.baseURL + "/user/list?groupId=" + groupId + "&userType=all&userStatus=all&limit=" + strconv.Itoa(limit) + offsetQueryParam

	req, err := client.newRequest(ctx, url, http.MethodGet, nil)
	if err != nil {
		return nil, fmt.Errorf("GET error creating list request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET list users failed: %w", err)
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
	url := client.baseURL + "/user?userId=" + user.UserID + "&groupId=" + user.GroupID

	req, err := client.newRequest(ctx, url, http.MethodDelete, nil)
	if err != nil {
		return fmt.Errorf("DELETE error creating request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE to cloudian /user got: %w", err)
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
	url := client.baseURL + "/group?groupId=" + groupId

	req, err := client.newRequest(ctx, url, http.MethodDelete, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE to cloudian /group got: %w", err)
	}

	return resp.Body.Close()
}

// Creates a group.
func (client Client) CreateGroup(ctx context.Context, group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := json.Marshal(toInternal(group))
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := client.newRequest(ctx, url, http.MethodPut, jsonData)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST to cloudian /group: %w", err)
	}

	return resp.Body.Close()
}

// Updates a group if it does not exists.
func (client Client) UpdateGroup(ctx context.Context, group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := json.Marshal(toInternal(group))
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a context with a timeout
	req, err := client.newRequest(ctx, url, http.MethodPost, jsonData)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("PUT to cloudian /group: %w", err)
	}

	return resp.Body.Close()
}

// Get a group. Returns an error even in the case of a group not found.
// This error can then be checked against ErrNotFound: errors.Is(err, ErrNotFound)
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

func (client Client) newRequest(ctx context.Context, url string, method string, body []byte) (*http.Request, error) {
	var buffer io.Reader = nil
	if body != nil {
		buffer = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, buffer)
	if err != nil {
		return req, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", client.authHeader)

	return req, nil
}
