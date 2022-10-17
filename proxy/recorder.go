package proxy

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type recorder struct {
	sync.Mutex
	responses map[string]entry
}

type entry struct {
	msg       dns.Msg
	timestamp time.Time
}

func newRecorder() *recorder {
	return &recorder{
		responses: make(map[string]entry),
	}
}

func (r *recorder) get(question *dns.Msg, maxAge time.Duration) (dns.Msg, bool) {
	r.Lock()
	defer r.Unlock()
	str := question.Question[0].String()

	found, ok := r.responses[str]
	if ok && time.Since(found.timestamp) < maxAge {
		return found.msg, true
	}
	return dns.Msg{}, false
}

func (r *recorder) add(question *dns.Msg, response *dns.Msg) {
	r.Lock()
	defer r.Unlock()
	str := question.Question[0].String()
	r.responses[str] = entry{*response, time.Now()}
}
