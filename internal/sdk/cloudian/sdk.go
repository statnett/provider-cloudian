package cloudian

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Group struct {
	Active             string
	GroupId            string
	GroupName          string
	LdapEnabled        bool
	LdapGroup          string
	LdapMatchAttribute string
	LdapSearch         string
	LdapSearchUserBase string
	LdapServerURL      string
	LdapUserDNTemplate string
	S3endpointshttp    []string
	S3endpointshttps   []string
	S3websiteendpoints []string
}

type User struct {
	UserId          string
	GroupId         string
	CanonicalUserId string
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

const baseUrl = "https://s3-admin.statnett.no:19443"

var client = &http.Client{}

// List all users of a group
func ListUsers(groupId string, offsetUserId *string, tokenBase64 string) ([]User, error) {
	var retVal []User

	limit := 100

	var offsetQueryParam string
	if offsetUserId == nil {
		offsetQueryParam = ""
	} else {
		offsetQueryParam = "&offset=" + *offsetUserId
	}

	url := baseUrl + "/user/list?groupId=" + groupId + "&userType=all&userStatus=all&limit=" + strconv.Itoa(limit) + offsetQueryParam

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("GET error creating list request: ", err)
		return nil, err
	}

	resp, err := client.Do(req)

	if err == nil {
		body, err := io.ReadAll(resp.Body)
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
			moreUsers, err := ListUsers(groupId, &users[limit].UserId, tokenBase64)

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
func DeleteUser(user User, tokenBase64 string) (*User, error) {
	url := baseUrl + "/user?userId=" + user.UserId + "&groupId=" + user.GroupId + "&canonicalUserId=" + user.CanonicalUserId

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("DELETE error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+tokenBase64)

	resp, err := client.Do(req)

	if resp != nil && err != nil {
		//Cloudian does not return a payload for this DELETE, but we can echo it to the callsite if all went well
		return &user, nil
	}
	return nil, err
}

// Delete a group and all its members
func DeleteGroupRecursive(groupId string, tokenBase64 string) (*string, error) {
	users, err := ListUsers(groupId, nil, tokenBase64)

	if err != nil {
		for _, user := range users {
			_, err := DeleteUser(user, tokenBase64)
			if err != nil {
				fmt.Println("Error deleting user: ", err)
				return nil, err
			}
		}

		retVal, err := DeleteGroup(groupId, tokenBase64)

		return retVal, err
	}

	return nil, err
}

// Deletes a group if it is without members
func DeleteGroup(groupId string, tokenBase64 string) (*string, error) {
	url := baseUrl + "/group?groupId=" + groupId

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println("DELETE error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+tokenBase64)

	resp, err := client.Do(req)

	statusErrStr := strconv.Itoa(resp.StatusCode)
	if err != nil {
		fmt.Println("DELETE to cloudian /group got status code ["+statusErrStr+"]", err)
		return nil, err
	}

	//Cloudian does not return a payload for this DELETE, but we can echo it to the callsite if all went well
	return &groupId, nil
}

func CreateGroup(group Group, tokenBase64 string) (*Group, error) {
	url := baseUrl + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("POST error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+tokenBase64)

	resp, err := client.Do(req)

	statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
	if err != nil {
		fmt.Println("POST to cloudian /group got status code ["+statusErrStr+"]", err)
		return nil, err
	}

	//Cloudian does not return a payload for this POST, but we can echo it to the callsite if all went well
	return &group, nil
}

func UpdateGroup(group Group, tokenBase64 string) (*Group, error) {
	url := baseUrl + "/group"

	jsonData, err := marshalGroup(group)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("POST error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+tokenBase64)

	resp, err := client.Do(req)

	statusErrStr := strconv.FormatInt(int64(resp.StatusCode), 10)
	if err != nil {
		fmt.Println("PUT to cloudian /group got status code ["+statusErrStr+"]", err)
	}

	//Cloudian does not return a payload for this PUT, but we can echo it to the callsite if all went well
	return &group, nil
}

func GetGroup(groupId string, tokenBase64 string) (*Group, error) {
	url := baseUrl + "/group?groupId=" + groupId

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("GET error creating request: ", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+tokenBase64)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("GET errored towards Cloudian /group: ", err)
		return nil, err
	}

	defer resp.Body.Close()

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
