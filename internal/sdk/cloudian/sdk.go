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

func MkClient(baseUrl string, tokenBase64 string) *Client {
	return &Client{
		baseURL:    baseUrl,
		httpClient: &http.Client{},
		token:      tokenBase64,
	}
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
		fmt.Println("GET error creating list request: ", err)
		return nil, err
	}

	resp, err := client.httpClient.Do(req)

	if err == nil {
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close() // nolint:errcheck
		if err != nil {
			fmt.Println("GET reading list users response body failed: ", err)
			return nil, err
		}

		users, err := unmarshalUsersJson(body)
		if err != nil {
			fmt.Println("GET unmarshal users response body failed: ", err)
			return nil, err
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
		fmt.Println("GET list users failed: ", err)
		return nil, err
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
		fmt.Println("DELETE error creating request: ", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

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
		for _, user := range users {
			err := client.DeleteUser(user)
			if err != nil {
				fmt.Println("Error deleting user: ", err)
				return err
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
		fmt.Println("DELETE error creating request: ", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.Itoa(resp.StatusCode)
		fmt.Println("DELETE to cloudian /group got status code ["+statusErrStr+"]", err)
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	return nil
}

func (client Client) CreateGroup(group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("POST error creating request: ", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
		fmt.Println("POST to cloudian /group got status code ["+statusErrStr+"]", err)
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	return err
}

func (client Client) UpdateGroup(group Group) error {
	url := client.baseURL + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("POST error creating request: ", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
		fmt.Println("PUT to cloudian /group got status code ["+statusErrStr+"]", err)
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
		fmt.Println("GET error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+client.token)

	resp, err := client.httpClient.Do(req)

	if err != nil {
		fmt.Println("GET errored towards Cloudian /group: ", err)
		return nil, err
	}

	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("GET reading response body failed: ", err)
			return nil, err
		}

		group, err := unmarshalGroupJson(body)
		if err != nil {
			fmt.Println("GET unmarshal response body failed: ", err)
			return nil, err
		}

		return &group, nil
	}

	// Cloudian-API returns 204 if the group does not exist
	return nil, nil
}
