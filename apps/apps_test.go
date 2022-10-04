package apps

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

const appFixture string = `
{
    "created": "2014-01-01T00:00:00UTC",
    "id": "example-go",
    "owner": "test",
    "structure": {},
    "updated": "2014-01-01T00:00:00UTC",
    "uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
}`

const appsFixture string = `
{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "created": "2014-01-01T00:00:00UTC",
            "id": "example-go",
            "owner": "test",
            "structure": {},
            "updated": "2014-01-01T00:00:00UTC",
            "uuid": "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75"
        }
    ]
}`

const appCreateExpected string = `{"id":"example-go"}`
const appRunExpected string = `{"command":"echo hi"}`
const appTransferExpected string = `{"owner":"test"}`

type fakeHTTPServer struct {
	createID        bool
	createWithoutID bool
}

func (f *fakeHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("DEIS_API_VERSION", deis.APIVersion)

	if req.URL.Path == "/v2/apps/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) == appCreateExpected && !f.createID {
			f.createID = true
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte(appFixture))
			return
		} else if string(body) == "" && !f.createWithoutID {
			f.createWithoutID = true
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte(appFixture))
			return
		}

		fmt.Printf("Unexpected Body: %s'\n", body)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/apps/" && req.Method == "GET" {
		res.Write([]byte(appsFixture))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/" && req.Method == "GET" {
		res.Write([]byte(appFixture))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/" && req.Method == "DELETE" {
		res.WriteHeader(http.StatusNoContent)
		res.Write(nil)
		return
	}

	// The entire log message is prefixed and suffixed with a few characters (not entirely sure why)
	// We mimic those here
	if req.URL.Path == "/v2/apps/example-go/logs" && req.URL.RawQuery == "" && req.Method == "GET" {
		res.Write([]byte(`test\nfoo\nbar`))
		return
	}

	// The entire log message is prefixed and suffixed with a few characters (not entirely sure why)
	// We mimic those here
	if req.URL.Path == "/v2/apps/example-go/logs" && req.URL.RawQuery == "log_lines=1" && req.Method == "GET" {
		res.Write([]byte("test"))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/run" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != appRunExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", appRunExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.Write([]byte(`{"exit_code":0,"output":"hi\n"}`))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != appTransferExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", appTransferExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.WriteHeader(http.StatusNoContent)
		res.Write(nil)
		return
	}

	fmt.Printf("Unrecongized URL %s\n", req.URL)
	res.WriteHeader(http.StatusNotFound)
	res.Write(nil)
}

func TestAppsCreate(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{createID: false, createWithoutID: false}
	server := httptest.NewServer(&handler)
	defer server.Close()

	expected := api.App{
		ID:      "example-go",
		Created: "2014-01-01T00:00:00UTC",
		Owner:   "test",
		Updated: "2014-01-01T00:00:00UTC",
		UUID:    "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75",
	}

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"example-go", ""} {
		actual, err := New(deis, id)

		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %v, Got %v", expected, actual)
		}
	}
}

func TestAppsGet(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	expected := api.App{
		ID:      "example-go",
		Created: "2014-01-01T00:00:00UTC",
		Owner:   "test",
		Updated: "2014-01-01T00:00:00UTC",
		UUID:    "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75",
	}

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := Get(deis, "example-go")

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, Got %v", expected, actual)
	}
}

func TestAppsDestroy(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = Delete(deis, "example-go"); err != nil {
		t.Fatal(err)
	}
}

func TestAppsRun(t *testing.T) {
	t.Parallel()

	expected := api.AppRunResponse{
		Output:     "hi\n",
		ReturnCode: 0,
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := Run(deis, "example-go", "echo hi")

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, Got %v", expected, actual)
	}
}

func TestAppsList(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	expected := api.Apps{
		{
			ID:      "example-go",
			Created: "2014-01-01T00:00:00UTC",
			Owner:   "test",
			Updated: "2014-01-01T00:00:00UTC",
			UUID:    "de1bf5b5-4a72-4f94-a10c-d2a3741cdf75",
		},
	}

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, _, err := List(deis, 100)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, Got %v", expected, actual)
	}
}

type testExpected struct {
	Input    int
	Expected string
}

func TestAppsLogs(t *testing.T) {
	t.Parallel()

	tests := []testExpected{
		{
			Input:    -1,
			Expected: `test\nfoo\nbar`,
		},
		{
			Input:    1,
			Expected: "test",
		},
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		actual, err := Logs(deis, "example-go", test.Input)

		if err != nil {
			t.Error(err)
		}

		if actual != test.Expected {
			t.Errorf("Expected %s, Got %s", test.Expected, actual)
		}
	}
}

func TestAppsTransfer(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = Transfer(deis, "example-go", "test"); err != nil {
		t.Fatal(err)
	}
}
