// Package config provides methods for managing configuration of apps.
package sharedvolumes

import (
	"encoding/json"
	"fmt"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
	"github.com/trilogy-group/devgraph-eyk-controller-sdk-go/api"
)

// List List shared volumes for a volume
func List(c *deis.Client, appID string, results int) (api.SharedVolumes, int, error) {
	u := fmt.Sprintf("/v2/apps/%s/sharedvolumes/", appID)
	body, count, reqErr := c.LimitedRequest(u, results)
	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return []api.SharedVolume{}, -1, reqErr
	}
	var sharedvolumes []api.SharedVolume
	if err := json.Unmarshal([]byte(body), &sharedvolumes); err != nil {
		return []api.SharedVolume{}, -1, err
	}
	return sharedvolumes, count, reqErr
}

// Create Create a shared Volume.
func Create(c *deis.Client, appID string, sharedVolume api.SharedVolume) (api.SharedVolume, error) {
	body, err := json.Marshal(sharedVolume)
	if err != nil {
		return api.SharedVolume{}, err
	}
	u := fmt.Sprintf("/v2/apps/%s/sharedvolumes/", appID)
	res, reqErr := c.Request("POST", u, body)
	if reqErr != nil {
		return api.SharedVolume{}, reqErr
	}
	defer res.Body.Close()
	newSharedVolume := api.SharedVolume{}
	if err = json.NewDecoder(res.Body).Decode(&newSharedVolume); err != nil {
		return api.SharedVolume{}, err
	}
	return newSharedVolume, reqErr
}

// Delete delete an app's shared volume.
func Delete(c *deis.Client, appID string, volumeID string) error {
	u := fmt.Sprintf("/v2/apps/%s/sharedvolumes/%s/", appID, volumeID)
	res, err := c.Request("DELETE", u, nil)
	if err == nil {
		res.Body.Close()
	}
	return err
}

// Mount mount an app's volume and creates a new release.
// This is a patching operation, which means when you call Mount() with an api.Volumes:
//
//    - If the variable does not exist, it will be set.
//    - If the variable exists, it will be overwritten.
//    - If the variable is set to nil, it will be unmount.
//    - If the variable was ignored in the api.Volumes, it will remain unchanged.
//
// Calling Mount() with an empty api.Volume will return a deis.ErrConflict.
// Trying to Unmount a key that does not exist returns a deis.ErrUnprocessable.
func Mount(c *deis.Client, appID string, name string, volume api.SharedVolume) (api.SharedVolume, error) {
	body, err := json.Marshal(volume)
	if err != nil {
		return api.SharedVolume{}, err
	}
	u := fmt.Sprintf("/v2/apps/%s/sharedvolumes/%s/path", appID, name)
	res, reqErr := c.Request("PATCH", u, body)
	if reqErr != nil {
		return api.SharedVolume{}, reqErr
	}
	defer res.Body.Close()
	newSharedVolume := api.SharedVolume{}
	if err = json.NewDecoder(res.Body).Decode(&newSharedVolume); err != nil {
		return api.SharedVolume{}, err
	}
	return newSharedVolume, reqErr
}
