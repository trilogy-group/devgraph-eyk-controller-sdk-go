// Package services provides methods for managing an app's services.
package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/controller-sdk-go/api"
)

// List services registered with an app.
func List(c *deis.Client, appID string) (api.Services, error) {
	u := fmt.Sprintf("/v2/apps/%s/services/", appID)
	res, reqErr := c.Request("GET", u, nil)

	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return []api.Service{}, reqErr
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []api.Service{}, err
	}

	r := make(map[string]interface{})
	if err = json.Unmarshal([]byte(body), &r); err != nil {
		return []api.Service{}, err
	}

	out, err := json.Marshal(r["services"].([]interface{}))
	if err != nil {
		return []api.Service{}, err
	}

	var services []api.Service
	if err := json.Unmarshal([]byte(out), &services); err != nil {
		return []api.Service{}, err
	}

	return services, reqErr
}

// New adds a service to an app.
func New(c *deis.Client, appID string, procfileType string, pathPattern string) (api.Service, error) {
	u := fmt.Sprintf("/v2/apps/%s/services/", appID)

	req := api.ServiceCreateUpdateRequest{ProcfileType: procfileType, PathPattern: pathPattern}

	body, err := json.Marshal(req)

	if err != nil {
		return api.Service{}, err
	}

	res, reqErr := c.Request("POST", u, body)
	if reqErr != nil && !deis.IsErrAPIMismatch(reqErr) {
		return api.Service{}, reqErr
	}
	defer res.Body.Close()

	d := api.Service{ProcfileType: procfileType, PathPattern: pathPattern}
	return d, reqErr
}

// Delete service from app
func Delete(c *deis.Client, appID string, procfileType string) error {
	u := fmt.Sprintf("/v2/apps/%s/services/", appID)

	req := api.ServiceDeleteRequest{ProcfileType: procfileType}

	body, err := json.Marshal(req)

	if err != nil {
		return err
	}

	_, err = c.Request("DELETE", u, body)

	return err
}
