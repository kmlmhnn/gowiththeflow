package counter

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/publicsuffix"

	g "github.com/kmlmhnn/gowiththeflow"
)

func counter(c chan struct{}, rs chan *g.Request) {
	defer close(c)
	for count := 0; count < 5; count++ {
		r := <-rs
		fmt.Fprintf(r, "%d", count)
		r.Done()
	}
}

func TestCounter(t *testing.T) {
	handler := g.Handler(counter)
	rand.Seed(time.Now().UnixNano())
	server := httptest.NewServer(handler)
	defer server.Close()
	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	f := func() {
		defer wg.Done()
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{
			Jar: jar,
		}
		var resp *http.Response
		for i := 0; i < 5; i++ {
			resp, err = client.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("(status) want: 200, got: %d\n", resp.StatusCode)
			}
			expected := fmt.Sprintf("%d", i)
			actual, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if expected != string(actual) {
				t.Fatalf("(body) want: %s, got: %s", expected, actual)
			}

		}
		resp, err = client.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("(status) want: 200, got: %d\n", resp.StatusCode)
		}
		expected := "0"
		actual, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if expected != string(actual) {
			t.Fatalf("(body) want: %s, got: %s", expected, actual)
		}
	}
	for i := 0; i < n; i++ {
		go f()
	}
	wg.Wait()
}
