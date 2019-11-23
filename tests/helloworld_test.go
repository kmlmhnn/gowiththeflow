package helloworld

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	g "github.com/kmlmhnn/gowiththeflow"
)

func helloworld(c chan struct{}, rs chan *g.Request) {
	for {
		r := <-rs
		fmt.Fprintf(r, "hello world")
		r.Done()
	}
}

func Test(t *testing.T) {
	handler := g.Handler(helloworld)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("(status) want: 200, got: %d\n", resp.StatusCode)
	}
	expected := "hello world"
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if expected != string(actual) {
		t.Fatalf("(body) want: %s, got: %s", expected, actual)
	}
}
