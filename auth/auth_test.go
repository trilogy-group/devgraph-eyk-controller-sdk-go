package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	deis "github.com/trilogy-group/devgraph-eyk-controller-sdk-go"
)

const registerExpected string = `{"username":"test","password":"opensesame","email":"test@example.com"}`
const loginExpected string = `{"username":"test","password":"opensesame"}`
const passwdExpected string = `{"username":"test","password":"old","new_password":"new"}`
const regenAllExpected string = `{"all":true}`
const regenUserExpected string = `{"username":"test"}`
const cancelUserExpected string = `{"username":"foo"}`
const cancelAdminExpected string = `{"username":"admin"}`

type fakeHTTPServer struct {
	regenBodyEmpty    bool
	regenBodyAll      bool
	regenBodyUsername bool
	cancelEmpty       bool
	cancelUsername    bool
}

func (f *fakeHTTPServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("DEIS_API_VERSION", deis.APIVersion)

	if req.URL.Path == "/v2/auth/register/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != registerExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", registerExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.WriteHeader(http.StatusCreated)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/auth/login/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		if string(body) != loginExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", loginExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.Write([]byte(`{"token":"abc"}`))
		return
	}

	if req.URL.Path == "/v2/auth/passwd/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) != passwdExpected {
			fmt.Printf("Expected '%s', Got '%s'\n", passwdExpected, body)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
			return
		}

		res.Write(nil)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if req.URL.Path == "/v2/auth/tokens/" && req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) == regenAllExpected && !f.regenBodyAll {
			f.regenBodyAll = true
			res.Write(nil)
			return
		} else if string(body) == regenUserExpected && !f.regenBodyUsername {
			f.regenBodyUsername = true
			res.Write([]byte(`{"token":"123"}`))
			return
		} else if string(body) == "" && !f.regenBodyEmpty {
			f.regenBodyEmpty = true
			res.Write([]byte(`{"token":"abc"}`))
			return
		}

		fmt.Printf("%s is not a valid body.", body)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(nil)
		return
	}

	if req.URL.Path == "/v2/auth/cancel/" && req.Method == "DELETE" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			fmt.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write(nil)
		}

		if string(body) == cancelAdminExpected && !f.cancelUsername {
			f.cancelUsername = true
			res.WriteHeader(http.StatusConflict)
			res.Write([]byte(`{"detail":"foo still has applications assigned. Delete or transfer ownership"}`))
			return
		} else if string(body) == cancelUserExpected && !f.cancelUsername {
			f.cancelUsername = true
			res.WriteHeader(http.StatusNoContent)
			res.Write(nil)
			return
		} else if string(body) == "" && !f.cancelEmpty {
			f.cancelEmpty = true
			res.WriteHeader(http.StatusNoContent)
			res.Write(nil)
			return
		}

		fmt.Printf("%s is not a valid body.", body)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write(nil)
		return
	}

	fmt.Printf("Unrecognized URL %s\n", req.URL)
	res.WriteHeader(http.StatusNotFound)
	res.Write(nil)
}

func TestRegister(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err = Register(deis, "test", "opensesame", "test@example.com"); err != nil {
		t.Error(err)
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	actual, err := Login(deis, "test", "opensesame")

	if err != nil {
		t.Error(err)
	}

	expected := "abc"

	if actual != expected {
		t.Errorf("Expected %s, Got %s", expected, actual)
	}
}

func TestPasswd(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err := Passwd(deis, "test", "old", "new"); err != nil {
		t.Error(err)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{cancelUsername: false, cancelEmpty: false}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(deis, "foo"); err != nil {
		t.Error(err)
	}

	if err := Delete(deis, ""); err != nil {
		t.Error(err)
	}
}

func TestDeleteUserApp(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{cancelUsername: false, cancelEmpty: false}
	server := httptest.NewServer(&handler)
	defer server.Close()

	d, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(d, "admin")
	// should be a 409 Conflict

	if err != deis.ErrCancellationFailed {
		t.Errorf("got '%s' but expected '%s'", err, deis.ErrConflict)
	}
}

func TestRegenerate(t *testing.T) {
	t.Parallel()

	handler := fakeHTTPServer{}
	server := httptest.NewServer(&handler)
	defer server.Close()

	deis, err := deis.New(false, server.URL, "abc")
	if err != nil {
		t.Fatal(err)
	}

	token, err := Regenerate(deis, "", true)

	if err != nil {
		t.Error(err)
	}

	if token != "" {
		t.Errorf("Expected token be empty, Got %s", token)
	}

	token, err = Regenerate(deis, "test", false)

	if err != nil {
		t.Error(err)
	}

	expected := "123"
	if token != expected {
		t.Errorf("Expected %s, Got %s", expected, token)
	}

	token, err = Regenerate(deis, "", false)

	if err != nil {
		t.Error(err)
	}

	expected = "abc"
	if token != expected {
		t.Errorf("Expected %s, Got %s", expected, token)
	}
}
