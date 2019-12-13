package gowiththeflow

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

type flow func(chan struct{}, chan *Request)

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

func Handler(f flow) http.HandlerFunc {
	q := make(chan *Request)
	go manage(q, f)
	return func(w http.ResponseWriter, r *http.Request) {
		var sid string
		if cookie, err := r.Cookie("sid"); err != nil {
			sid = ""
		} else {
			sid = cookie.Value
		}
		req := &Request{
			Request: r,
			sid:     sid,
			res: &response{
				header: make(http.Header, 0),
				body:   []byte{},
			},
			done: make(chan struct{}),
		}
		q <- req
		<-req.done

		res := req.res
		if req.sid != res.sid {
			http.SetCookie(w, &http.Cookie{
				Name:  "sid",
				Value: res.sid,
			})
		}
		for key, values := range res.header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		if res.status != 0 {
			w.WriteHeader(res.status)
		}
		w.Write(res.body)
	}
}

func manage(q chan *Request, f flow) {
	type session struct {
		rchan chan *Request
		cchan chan struct{}
	}

	sessions := make(map[string]session)

	newsession := func() (string, session) {
		rid := make([]byte, 20)
		rand.Read(rid)
		sid := base64.StdEncoding.EncodeToString(rid)
		rc := make(chan *Request)
		cc := make(chan struct{})
		return sid, session{rchan: rc, cchan: cc}
	}

	for {
		req := <-q
		res := req.res
		exists := false
		if s, valid := sessions[req.sid]; valid {
			select {
			case <-s.cchan:
				delete(sessions, req.sid)
			default:
				exists = true
				res.sid = req.sid
				go func() { s.rchan <- req }()
			}
		}
		if !exists {
			sid, s := newsession()
			sessions[sid] = s
			res.sid = sid
			go f(s.cchan, s.rchan)
			go func() { s.rchan <- req }()
		}
	}
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
