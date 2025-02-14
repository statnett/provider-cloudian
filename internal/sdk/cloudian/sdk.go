package cloudian

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

const ListLimit = 100

type Client struct {
	client *resty.Client
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
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

var ErrNotFound = errors.New("not found")

// WithInsecureTLSVerify skips the TLS validation of the server certificate when `insecure` is true.
func WithInsecureTLSVerify(insecure bool) func(*Client) {
	return func(c *Client) {
		c.client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: insecure}) //nolint:gosec
	}
}

func NewClient(baseURL string, authHeader string, opts ...func(*Client)) *Client {
	c := &Client{
		client: resty.New().
			SetBaseURL(baseURL).
			SetHeader("Authorization", authHeader),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// List all users of a group.
func (client Client) ListUsers(ctx context.Context, groupID string, userID *string) ([]User, error) {
	params := map[string]string{
		"groupId":    groupID,
		"userType":   "all",
		"userStatus": "all",
		"limit":      strconv.Itoa(ListLimit),
	}
	if userID != nil {
		params["offset"] = *userID
	}

	var users []User
	_, err := client.newRequest(ctx).
		SetQueryParams(params).
		SetResult(&users).
		Get("/user/list")
	if err != nil {
		return nil, fmt.Errorf("GET list users failed: %w", err)
	}

	// Paginated API endpoint where limit+1 elements indicates more pages
	if len(users) > ListLimit {
		// Fetch remaining users starting from the user after the limit
		moreUsers, err := client.ListUsers(ctx, groupID, &users[ListLimit].UserID)
		if err != nil {
			return nil, err
		}
		// Exclude first user as it's a duplicate from previous fetch
		users = append(users, moreUsers[1:]...)
	}

	return users, nil
}

// Delete a single user. Errors if the user does not exist.
func (client Client) DeleteUser(ctx context.Context, user User) error {
	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{
			"groupId": user.GroupID,
			"userId":  user.UserID,
		}).
		Delete("/user")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE user unexpected status: %d", resp.StatusCode())
	}

}

// Create a single user of type `User` into a groupId
func (client Client) CreateUser(ctx context.Context, user User) error {
	resp, err := client.newRequest(ctx).
		SetBody(toInternalUser(user)).
		Put("/user")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("CREATE user unexpected status: %d", resp.StatusCode())
	}
}

// GetUser gets a user. Returns an error even in the case of a user not found.
// This error can then be checked against ErrNotFound: errors.Is(err, ErrNotFound)
func (client Client) GetUser(ctx context.Context, user User) (*User, error) {
	// FIXME: Introduce UserId struct and enrich User
	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{
			"groupId": user.GroupID,
			"userId":  user.UserID,
		}).
		Get("/user")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		return &user, nil
	case 204:
		// Cloudian-API returns 204 if the user does not exist
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error: GET user unexpected status code: %d", resp.StatusCode())
	}
}

// CreateUserCredentials creates a new set of credentials for a user.
func (client Client) CreateUserCredentials(ctx context.Context, user User) (*SecurityInfo, error) {
	var securityInfo SecurityInfo

	resp, err := client.newRequest(ctx).
		SetResult(&securityInfo).
		SetQueryParams(map[string]string{"groupId": user.GroupID, "userId": user.UserID}).
		Put("/user/credentials")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		return &securityInfo, nil
	default:
		return nil, fmt.Errorf("CREATE user credentials unexpected status: %d", resp.StatusCode())
	}
}

// GetUserCredentials fetches all the credentials of a user.
func (client Client) GetUserCredentials(ctx context.Context, accessKey string) (*SecurityInfo, error) {
	var securityInfo SecurityInfo

	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{"accessKey": accessKey}).
		SetResult(&securityInfo).
		Get("/user/credentials")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		return &securityInfo, nil
	case 204:
		// Cloudian-API returns 204 if no security credentials found
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error: list credentials unexpected status code: %d", resp.StatusCode())
	}
}

// ListUserCredentials fetches all the credentials of a user.
func (client Client) ListUserCredentials(ctx context.Context, user User) ([]SecurityInfo, error) {
	var securityInfo []SecurityInfo

	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{"groupId": user.GroupID, "userId": user.UserID}).
		SetResult(&securityInfo).
		Get("/user/credentials/list")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		return securityInfo, nil
	case 204:
		// Cloudian-API returns 204 if no security credentials found
		return nil, nil
	default:
		return nil, fmt.Errorf("error: list credentials unexpected status code: %d", resp.StatusCode())
	}
}

// DeleteUserCredentials deletes a set of credentials for a user.
func (client Client) DeleteUserCredentials(ctx context.Context, accessKey string) error {
	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{"accessKey": accessKey}).
		Delete("/user/credentials")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE credentials unexpected status: %d", resp.StatusCode())
	}
}

// Delete a group and all its members.
func (client Client) DeleteGroupRecursive(ctx context.Context, groupID string) error {
	users, err := client.ListUsers(ctx, groupID, nil)
	if err != nil {
		return fmt.Errorf("error listing users: %w", err)
	}

	for _, user := range users {
		if err := client.DeleteUser(ctx, user); err != nil {
			return fmt.Errorf("error deleting user: %w", err)
		}
	}

	return client.DeleteGroup(ctx, groupID)
}

// Deletes a group if it is without members.
func (client Client) DeleteGroup(ctx context.Context, groupID string) error {
	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{"groupId": groupID}).
		Delete("/group")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return nil
	default:
		return fmt.Errorf("DELETE group unexpected statusCode: %d", resp.StatusCode())
	}
}

// Creates a group.
func (client Client) CreateGroup(ctx context.Context, group Group) error {
	resp, err := client.newRequest(ctx).
		SetBody(toInternal(group)).
		Put("/group")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return err
	default:
		return fmt.Errorf("CREATE group unexpected status: %d", resp.StatusCode())
	}
}

// Updates a group if it does not exists.
func (client Client) UpdateGroup(ctx context.Context, group Group) error {
	resp, err := client.newRequest(ctx).
		SetBody(toInternal(group)).
		Post("/group")
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case 200:
		return err
	default:
		return fmt.Errorf("Update group unexpected status: %d", resp.StatusCode())
	}
}

// Get a group. Returns an error even in the case of a group not found.
// This error can then be checked against ErrNotFound: errors.Is(err, ErrNotFound)
func (client Client) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	var group groupInternal
	resp, err := client.newRequest(ctx).
		SetQueryParams(map[string]string{"groupId": groupID}).
		SetResult(&group).
		Get("/group")
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case 200:
		retVal := fromInternal(group)
		return &retVal, nil
	case 204:
		// Cloudian-API returns 204 if the group does not exist
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error: GET group unexpected status code: %d", resp.StatusCode())
	}
}

func (client Client) newRequest(ctx context.Context) *resty.Request {
	return client.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		ForceContentType("application/json") // TODO figure out why this is needed
}
