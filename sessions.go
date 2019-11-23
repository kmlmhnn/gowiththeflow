package gowiththeflow

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

type session struct {
	rchan chan *Request
	cchan chan struct{}
}

func newsession() (string, session) {
	rid := make([]byte, 20)
	if _, err := rand.Read(rid); err != nil {
		log.Fatalf("rand.Read failed: %v\n", err)
	}
	sid := base64.StdEncoding.EncodeToString(rid)
	rc := make(chan *Request)
	cc := make(chan struct{})
	return sid, session{rchan: rc, cchan: cc}

}

func manage(q chan *Request, f flow) {
	sessions := make(map[string]session)
	for {
		req := <-q
		exists := false
		if s, valid := sessions[req.sid]; valid {
			select {
			case <-s.cchan:
				delete(sessions, req.sid)
			default:
				exists = true
				req.res.sid = req.sid
				go func() { s.rchan <- req }()
			}
		}
		if !exists {
			sid, s := newsession()
			sessions[sid] = s
			req.res.sid = sid
			go f(s.cchan, s.rchan)
			go func() { s.rchan <- req }()
		}
	}
}
