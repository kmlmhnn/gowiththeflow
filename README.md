# gowiththeflow

Invert the inversion of control using channels and goroutines.

## Example
```go
package main

import (
	"fmt"
	"net/http"

	g "github.com/kmlmhnn/gowiththeflow"
)

func count(c chan struct{}, rs chan *g.Request) {
	defer close(c)

	for counter := 0; counter < 5; counter++ {
		r := <-rs
		fmt.Fprintf(r, "%d\n", counter)
		r.Done()
	}
}

func main() {
	server := &http.Server{Addr: ":3000"}
	handler := g.Handler(count)
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

## Testing
```bash
$ echo tests/* | xargs -n 1 go test

```

## License

Copyright (C) 2019 Kamal M Nair & other gowiththeflow contributors

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
