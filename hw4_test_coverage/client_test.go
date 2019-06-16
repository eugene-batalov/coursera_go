package main

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"time"
	"io"
	"encoding/xml"
	"os"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

// код писать тут

func SearchServer(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	file, _ := os.Open("dataset.xml")

	fileContents, _ := ioutil.ReadAll(file)
	data := Root{}
	xml.Unmarshal(fileContents, &data)
	limit, _ := r.URL.Query()["limit"]
	offset, _ := r.URL.Query()["offset"]
	query, _ := r.URL.Query()["query"]
	order_field, _ := r.URL.Query()["order_field"]
	order_by, _ := r.URL.Query()["order_by"]
	fmt.Println(limit)
	bytes, _ := json.Marshal(data.Row[0:2])
	s := string(bytes)
	io.WriteString(w, s)
}

type Root struct {
	XMLName xml.Name `xml:"root"`
	Text    string   `xml:",chardata"`
	Row     []struct {
		Text          string `xml:",chardata"`
		ID            int    `xml:"id"`
		Guid          string `xml:"guid"`
		IsActive      string `xml:"isActive"`
		Balance       string `xml:"balance"`
		Picture       string `xml:"picture"`
		Age           int    `xml:"age"`
		EyeColor      string `xml:"eyeColor"`
		FirstName     string `xml:"first_name"`
		LastName      string `xml:"last_name"`
		Gender        string `xml:"gender"`
		Company       string `xml:"company"`
		Email         string `xml:"email"`
		Phone         string `xml:"phone"`
		Address       string `xml:"address"`
		About         string `xml:"about"`
		Registered    string `xml:"registered"`
		FavoriteFruit string `xml:"favoriteFruit"`
	} `xml:"row"`
}

var (
	cases = []SearchRequest{
		SearchRequest{
			Limit: -1,
		},
		SearchRequest{
			Limit:  50,
			Offset: -1,
		},
		SearchRequest{
			Limit: 17,
		},
		SearchRequest{
			Limit: 2,
		},
	}
	searchClient = &SearchClient{
		AccessToken: "TestAccessToken",
	}
)

func TestFindUsersBadLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	response, err := searchClient.FindUsers(cases[0])
	if response != nil && err.Error() != "limit must be > 0" {
		t.Error("limit < 0 should produce error")
	}
	ts.Close()
}

func TestFindUsersBadOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	response, err := searchClient.FindUsers(cases[1])
	if response != nil || err.Error() != "offset must be > 0" {
		t.Error("offset < 0 should produce error")
	}
	ts.Close()
}

func TestFindUsersTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1100 * time.Millisecond)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "timeout for limit=18&offset=0&order_by=0&order_field=&query=" {
		t.Error("should produce timeout error")
	}
	ts.Close()
}

func TestFindUsersUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	searchClient.URL = "BADURL"
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "unknown error Get BADURL?limit=18&offset=0&order_by=0&order_field=&query"+
		"=: unsupported protocol scheme \"\"" {
		t.Error("should produce unknown error")
	}
	ts.Close()
}

func TestFindUsersStatusUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "")
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "Bad AccessToken" {
		t.Error("should produce Bad AccessToken")
	}
	ts.Close()
}

func TestFindUsersStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "")
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "SearchServer fatal error" {
		t.Error("should produce SearchServer fatal error")
	}
	ts.Close()
}

func TestFindUsersStatusReturnStatusBadRequest1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "")
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "cant unpack error json: unexpected end of JSON input" {
		t.Error("should produce cant unpack error json: ")
	}
	ts.Close()
}

func TestFindUsersStatusReturnStatusBadRequest2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"ErrorBadOrderField"}`)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "OrderFeld  invalid" {
		t.Error("should produce OrderField invalid")
	}
	ts.Close()
}

func TestFindUsersStatusReturnStatusBadRequest3(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"Error"}`)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "unknown bad request error: Error" {
		t.Error("should produce unknown bad request error: Error")
	}
	ts.Close()
}

func TestFindUsersStatusReturnBadJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"Error":"Error"}`)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[2])
	if response != nil || err.Error() != "cant unpack result json: json: cannot unmarshal object into Go value of type []main.User" {
		t.Error("should produce cant unpack result json:")
	}
	ts.Close()
}

func TestFindUsersStatusReturnLimitEquals(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		file, _ := os.Open("dataset.xml")

		fileContents, _ := ioutil.ReadAll(file)
		data := Root{}
		xml.Unmarshal(fileContents, &data)
		bytes, _ := json.Marshal(data.Row[0:3])
		s := string(bytes)
		io.WriteString(w, s)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[3])
	if response != nil || err.Error() != "cant unpack result json: json: cannot unmarshal object into Go value of type []main.User" {
		t.Error("should produce cant unpack result json:")
	}
	ts.Close()
}

func TestFindUsersStatusReturnLimitLess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		file, _ := os.Open("dataset.xml")

		fileContents, _ := ioutil.ReadAll(file)
		data := Root{}
		xml.Unmarshal(fileContents, &data)
		bytes, _ := json.Marshal(data.Row[0:2])
		s := string(bytes)
		io.WriteString(w, s)
	}))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[3])
	if response != nil || err.Error() != "cant unpack result json: json: cannot unmarshal object into Go value of type []main.User" {
		t.Error("should produce cant unpack result json:")
	}
	ts.Close()
}

func TestFindUsersStatusUseSearchServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	searchClient.URL = ts.URL
	response, err := searchClient.FindUsers(cases[3])
	if response != nil || err.Error() != "cant unpack result json: json: cannot unmarshal object into Go value of type []main.User" {
		t.Error("should produce cant unpack result json:")
	}
	ts.Close()
}
