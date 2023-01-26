package perms

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
)

const adminFixture string = `
{
    "count": 1,
    "next": null,
    "previous": null,
    "results": [
        {
            "username": "test",
            "is_superuser": true
        },
        {
            "username": "foo",
            "is_superuser": true
        }
    ]
}`

const appPermsFixture string = `
{
  "users": [
    "test",
    "testing"
  ]
}`

const adminCreateExpected = `{"username":"test"}`
const appCreateExpected = `{"username":"foo"}`

type fakeHTTPServer struct{}

func (fakeHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("DEIS_API_VERSION", deis.APIVersion)

	if req.URL.Path == "/v2/admin/perms/" && req.Method == "GET" {
		res.Write([]byte(adminFixture))
		return
	}

	if req.URL.Path == "/v2/admin/perms/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != adminCreateExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", adminCreateExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.WriteHeader(http.StatusCreated)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/admin/perms/test" && req.Method == "DELETE" {
		res.WriteHeader(http.StatusNoContent)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/apps/example-go/perms/" && req.Method == "GET" {
		res.Write([]byte(appPermsFixture))
		return
	}

	if req.URL.Path == "/v2/apps/example-go/perms/foo" && req.Method == "DELETE" {
		res.WriteHeader(http.StatusNoContent)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/apps/example-go/perms/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != appCreateExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", appCreateExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.WriteHeader(http.StatusCreated)
		res.Write(nil)
		return
	}

	fmt.Printf("Unrecognized URL %s\n", req.URL)
	res.WriteHeader(http.StatusNotFound)
	res.Write(nil)
}

func TestList(t *testing.T) {
	t.Parallel()

	expected := []string{
		"test",
		"testing",
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := List(deis, "example-go")

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, Got %v", expected, actual)
	}
}

func TestListAdmins(t *testing.T) {
	t.Parallel()

	expected := []string{
		"test",
		"foo",
	}

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, _, err := ListAdmins(deis, 100)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, Got %v", expected, actual)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = New(deis, "example-go", "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestNewAdmin(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = NewAdmin(deis, "test"); err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = Delete(deis, "example-go", "foo"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAdmin(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = DeleteAdmin(deis, "test"); err != nil {
		t.Fatal(err)
	}
}
