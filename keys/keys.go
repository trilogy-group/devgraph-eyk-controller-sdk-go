// Package keys provides methods for managing a user's ssh keys.
package keys

import (
	"encoding/json"
	"fmt"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
	"github.com/trilogy-group/devgraph-eyk-controller-sdk-go/api"
)

// List lists a user's ssh keys.
func List(c *deis.Client, results int) (api.Keys, int, error) {
	body, count, reqErr := c.LimitedRequest("/v2/keys/", results)

	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return []api.Key{}, -1, reqErr
	}

	var keys []api.Key
	if err := json.Unmarshal([]byte(body), &keys); err != nil {
		return []api.Key{}, -1, err
	}

	return keys, count, reqErr
}

// New adds a new ssh key for the user. This is used for authenting with the git
// remote for the builder. This key must be unique to the current user, or the error
// deis.ErrDuplicateKey will be returned.
func New(c *deis.Client, id string, pubKey string) (api.Key, error) {
	req := api.KeyCreateRequest{ID: id, Public: pubKey}
	body, err := json.Marshal(req)
	if err != nil {
		return api.Key{}, err
	}

	res, reqErr := c.Request("POST", "/v2/keys/", body)
	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return api.Key{}, reqErr
	}
	defer res.Body.Close()

	key := api.Key{}
	if err = json.NewDecoder(res.Body).Decode(&key); err != nil {
		return api.Key{}, err
	}

	return key, reqErr
}

// Delete removes a user's ssh key. The key ID will be the key comment, usually the email or user@hostname
// of the user. The exact keyID can be retrieved with List()
func Delete(c *deis.Client, keyID string) error {
	u := fmt.Sprintf("/v2/keys/%s", keyID)

	res, err := c.Request("DELETE", u, nil)
	if err == nil {
		res.Body.Close()
	}
	return err
}
