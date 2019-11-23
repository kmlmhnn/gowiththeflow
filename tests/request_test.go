package req

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	g "github.com/kmlmhnn/gowiththeflow"
	"golang.org/x/net/publicsuffix"
)

func req(c chan struct{}, rs chan *g.Request) {
	defer close(c)

	r := <-rs
	n, _ := strconv.Atoi(r.FormValue("n"))
	r.WriteHeader(200 + n)
	r.Done()

	r = <-rs
	nick := r.FormValue("nick")
	r.Header().Set("nick", nick)
	r.Done()

	r = <-rs
	fmt.Fprintf(r, "hello %s", nick)
	r.Done()
}

func Test(t *testing.T) {
	handler := g.Handler(req)
	rand.Seed(time.Now().UnixNano())
	server := httptest.NewServer(handler)
	defer server.Close()
	var wg sync.WaitGroup
	n := 5
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
		x := rand.Intn(100)
		var resp *http.Response
		v := url.Values{}
		v.Set("n", fmt.Sprintf("%d", x))
		resp, err = client.PostForm(server.URL, v)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200+x {
			t.Fatalf("(status) want: %d, got: %d\n", 200+x, resp.StatusCode)
		}
		v.Set("nick", fmt.Sprintf("abc%d", x))
		resp, err = client.PostForm(server.URL, v)
		if resp.StatusCode != 200 {
			t.Fatalf("(status) want: 200, got: %d\n", resp.StatusCode)
		}
		if resp.Header.Get("nick") != v.Get("nick") {
			t.Fatalf("(header) want: %s, got: %s\n", v.Get("nick"), resp.Header.Get("nick"))
		}
		resp, err = client.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		expected := fmt.Sprintf("hello %s", v.Get("nick"))
		actual, err := ioutil.ReadAll(resp.Body)
		if expected != string(actual) {
			t.Fatalf("(body) want: %s, got: %s", expected, actual)
		}
	}
	for i := 0; i < n; i++ {
		go f()
	}
	wg.Wait()
}
