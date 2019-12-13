# gowiththeflow
Invert the inversion of control using channels and goroutines.

## Design
1. All incoming http requests are processed in the context of a session.
   If the session information (stored as a session cookie) is missing or invalid, a new session will be created to process that request.
2. Each session corresponds to a goroutine.
   Goroutines are spawned when sessions are created. When a goroutine returns, the corresponding session expires.
3. A function that will be executed as such a goroutine is called a `flow`.
   ```
   type flow func(chan struct{}, chan *Request)
   ```
   Closing `chan struct{}` signals the expiry of this session.
   All http requests that belong to this session can be received from `chan *Request`.

## Example 1
```go
package main

import (
	"fmt"
	"net/http"

	g "github.com/kmlmhnn/gowiththeflow"
)

// count is a flow that models a counter
func count(c chan struct{}, rs chan *g.Request) {
	defer close(c)                              // close c on return

	for counter := 0; counter < 5; counter++ {
		r := <-rs                           // get the next request

		fmt.Fprintf(r, "%d\n", counter)     // respond with the current value of counter

		r.Done()                            // we are done processing this request
	}
}

func main() {
	server := &http.Server{Addr: ":3000"}
	handler := g.Handler(count)                 // Handler transforms a flow to a normal handler
	http.HandleFunc("/count", handler)
	server.ListenAndServe()
}

```
``` bash
$ curl -c jar localhost:3000/count
0
$ curl -b jar localhost:3000/count
1
$ curl -b jar localhost:3000/count
2
$ curl -b jar localhost:3000/count
3
$ curl -b jar localhost:3000/count
4
$ curl -b jar localhost:3000/count
0
```

## Example 2
```go
func add(c chan struct{}, rs chan *g.Request) {
	defer close(c)

	a := 0
	for r := range rs {
		sa := r.FormValue("a")              // Request has an embedded *http.Request
		if sa != "" {
			a, _ = strconv.Atoi(sa)
			fmt.Fprintln(r, "ok")       // *Request implements http.ResponseWriter
			r.Done()
			break

		}
		r.WriteHeader(400)
		fmt.Fprintln(r, "bad request")
		r.Done()
	}

	b := 0
	for r := range rs {
		sb := r.FormValue("b")
		if sb != "" {
			b, _ = strconv.Atoi(sb)
			fmt.Fprintln(r, "ok")
			r.Done()
			break
		}
		r.WriteHeader(400)
		fmt.Fprintln(r, "bad request")
		r.Done()
	}

	r := <-rs
	fmt.Fprintf(r, "%d\n", a+b)
	r.Done()
}
```
``` bash
$ curl -c jar localhost:3000/
bad request
$ curl -b jar localhost:3000/?a=1
ok
$ curl -b jar localhost:3000/?a=2
bad request
$ curl -b jar localhost:3000/?b=2
ok
$ curl -b jar localhost:3000/
3
$ curl -b jar localhost:3000/
bad request
```

## Testing
```bash
$ echo tests/* | xargs -n 1 go test

```

## License
Copyright (C) 2019 Kamal M Nair

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
