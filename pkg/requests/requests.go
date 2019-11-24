package requests

import (
	"net/http"
)

// Переделать в интерфейс!!!
// Request Непосредственно запрос
type Request struct {
	http.Request
	http.Header
	H2   bool
	Body []byte
	Rate int
}

func NewRequest(url string, method string, http2 bool, header http.Header,
	body []byte, rate int) (*Request, error) {
	r := new(Request)
	if hr, e := http.NewRequest(method, url, nil); e == nil {
		r.Request = *hr
	} else {
		return nil, e
	}
	
	// deep copy of the Header
	r.Header = make(http.Header, len(header))
	for k, s := range header {
		r.Header[k] = append([]string(nil), s...)
	}

	r.H2 = http2
	r.Rate = rate

	return r, nil
}

func NewRequestGenerator(s [][]*Request) func() *Request {
	scripts := s
	sid := 0
	rid := 0
	count := s[sid][rid].Rate

	rg := func() *Request {
		if count == 0 {
			rid += 1
			if rid == len(scripts[sid]) {
				rid = 0
				sid += 1
			}
			if sid == len(scripts) {
				sid = 0
			}
			count = scripts[sid][rid].Rate
		}
		count -= 1
		return scripts[sid][rid]
	}

	return rg
}
