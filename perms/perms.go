// Package perms provides methods for managing user app and administrative permissions.
package perms

import (
	"encoding/json"
	"fmt"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
	"github.com/trilogy-group/devgraph-eyk-controller-sdk-go/api"
)

// List users that can access an app.
func List(c *deis.Client, appID string) ([]string, error) {
	res, reqErr := c.Request("GET", fmt.Sprintf("/v2/apps/%s/perms/", appID), nil)
	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return []string{}, reqErr
	}
	defer res.Body.Close()

	var users api.PermsAppResponse
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return []string{}, err
	}

	return users.Users, reqErr
}

// ListAdmins lists deis platform administrators.
func ListAdmins(c *deis.Client, results int) ([]string, int, error) {
	body, count, reqErr := c.LimitedRequest("/v2/admin/perms/", results)

	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return []string{}, -1, reqErr
	}

	var users []api.PermsRequest
	if err := json.Unmarshal([]byte(body), &users); err != nil {
		return []string{}, -1, err
	}

	usersList := []string{}

	for _, user := range users {
		usersList = append(usersList, user.Username)
	}

	return usersList, count, reqErr
}

// New gives a user access to an app.
func New(c *deis.Client, appID string, username string) error {
	return doNew(c, fmt.Sprintf("/v2/apps/%s/perms/", appID), username)
}

// NewAdmin makes a user an administrator.
func NewAdmin(c *deis.Client, username string) error {
	return doNew(c, "/v2/admin/perms/", username)
}

func doNew(c *deis.Client, u string, username string) error {
	req := api.PermsRequest{Username: username}

	reqBody, err := json.Marshal(req)

	if err != nil {
		return err
	}

	res, err := c.Request("POST", u, reqBody)
	if err == nil {
		res.Body.Close()
	}

	return err
}

// Delete removes a user from an app.
func Delete(c *deis.Client, appID string, username string) error {
	return doDelete(c, fmt.Sprintf("/v2/apps/%s/perms/%s", appID, username))
}

// DeleteAdmin removes administrative privileges from a user.
func DeleteAdmin(c *deis.Client, username string) error {
	return doDelete(c, fmt.Sprintf("/v2/admin/perms/%s", username))
}

func doDelete(c *deis.Client, u string) error {
	res, err := c.Request("DELETE", u, nil)
	if err == nil {
		res.Body.Close()
	}
	return err
}
