package gowiththeflow

import (
	"net/http"
)

type response struct {
	sid    string
	status int
	header http.Header
	body   []byte
}

type Request struct {
	*http.Request
	sid  string
	res  *response
	done chan struct{}
}

func (r *Request) Done() {
	close(r.done)
}

func (r *Request) Write(data []byte) (int, error) {
	r.res.body = append(r.res.body, data...)
	return len(r.res.body), nil

}

func (r *Request) Header() http.Header {
	return r.res.header
}

func (r *Request) WriteHeader(statusCode int) {
	r.res.status = statusCode
}
