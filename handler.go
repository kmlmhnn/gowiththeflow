package gowiththeflow

import (
	"net/http"
)

type flow func(chan struct{}, chan *Request)

func Handler(f flow) http.HandlerFunc {
	q := make(chan *Request)
	go manage(q, f)
	return handlerfn(q)
}

func handlerfn(q chan *Request) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var sid string
		if cookie, err := r.Cookie("gwtf_session"); err != nil {
			sid = ""
		} else {
			sid = cookie.Value
		}

		res := &response{"", 200, make(http.Header, 0), []byte{}}
		req := &Request{r, sid, res, make(chan struct{})}
		q <- req
		<-req.done

		if sid != req.res.sid {
			http.SetCookie(w, &http.Cookie{
				Name:  "gwtf_session",
				Value: req.res.sid,
			})
		}

		for k, _ := range res.header {
			w.Header().Add(k, res.header.Get(k))
		}
		w.WriteHeader(res.status)
		w.Write(res.body)
	}
	return handler
}
