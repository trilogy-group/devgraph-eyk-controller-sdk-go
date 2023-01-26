package builds

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
	"github.com/trilogy-group/devgraph-eyk-controller-sdk-go/api"
)

const buildsFixture string = `
{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "app": "example-go",
            "created": "2014-01-01T00:00:00UTC",
            "dockerfile": "FROM deis/slugrunner RUN mkdir -p /app WORKDIR /app ENTRYPOINT [\"/runner/init\"] ADD slug.tgz /app ENV GIT_SHA 060da68f654e75fac06dbedd1995d5f8ad9084db",
            "image": "example-go",
            "owner": "test",
            "procfile": {
                "web": "example-go"
            },
            "sha": "060da68f",
            "updated": "2014-01-01T00:00:00UTC",
            "uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
        }
    ]
}`

const buildFixture string = `
{
    "app": "example-go",
    "created": "2014-01-01T00:00:00UTC",
    "dockerfile": "",
    "image": "deis/example-go:latest",
    "owner": "test",
    "procfile": {
        "web": "example-go"
    },
    "sha": "",
    "updated": "2014-01-01T00:00:00UTC",
    "uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
}`

const buildExpected string = `{"image":"deis/example-go","procfile":{"web":"example-go"}}`

type fakeHTTPServer struct{}

func (fakeHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("DEIS_API_VERSION", deis.APIVersion)

	if req.URL.Path == "/v2/apps/example-go/builds/" && req.Method == "GET" {
		res.Write([]byte(buildsFixture))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/builds/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != buildExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", buildExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(buildFixture))
		return
	}

	fmt.Printf("Unrecognized URL %s\n", req.URL)
	res.WriteHeader(http.StatusNotFound)
	res.Write(nil)
}

func TestBuildsList(t *testing.T) {
	t.Parallel()

	expected := []api.Build{
		{
			App:        "example-go",
			Created:    "2014-01-01T00:00:00UTC",
			Dockerfile: "FROM deis/slugrunner RUN mkdir -p /app WORKDIR /app ENTRYPOINT [\"/runner/init\"] ADD slug.tgz /app ENV GIT_SHA 060da68f654e75fac06dbedd1995d5f8ad9084db",
			Image:      "example-go",
			Owner:      "test",
			Procfile: map[string]string{
				"web": "example-go",
			},
			Sha:     "060da68f",
			Updated: "2014-01-01T00:00:00UTC",
			UUID:    "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75",
		},
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, _, err := List(deis, "example-go", 100)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Error(fmt.Errorf("Expected %v, Got %v", expected, actual))
	}
}

func TestBuildCreate(t *testing.T) {
	t.Parallel()

	expected := api.Build{
		App:     "example-go",
		Created: "2014-01-01T00:00:00UTC",
		Image:   "deis/example-go:latest",
		Owner:   "test",
		Procfile: map[string]string{
			"web": "example-go",
		},
		Updated: "2014-01-01T00:00:00UTC",
		UUID:    "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75",
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	procfile := map[string]string{
		"web": "example-go",
	}

	actual, err := New(deis, "example-go", "deis/example-go", procfile)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Error(fmt.Errorf("Expected %v, Got %v", expected, actual))
	}
}
