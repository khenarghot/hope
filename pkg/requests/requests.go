package requests

import (
	"net/http"
)

// Переделать в интерфейс!!!
// Request Непосредственно запрос
type Request struct {
	http.Request
	http.Header
	Body []byte
	Rate int
	Skip int
}

type RequestGenerator func() *Request

func NewRequest(url string, method string, header http.Header,
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
	r.Rate = 1
	if rate > 0 {
		r.Rate = rate
	}
	r.Skip = 0

	return r, nil
}

func getStartRequest(sc [][]*Request, offset int) (s, r int) {
	if offset == 0 {
		return 0, 0
	}
	total := 0
	for _, sr := range sc {
		total += len(sr)
	}
	if total == 0 {
		panic("Empty request is bad idea")
	}
	offset = offset % total
	cur := 0
	s, r = 0, 0
	for _, sr := range sc {
		if len(sr)+cur <= offset {
			s++
			cur += len(sr)
		} else {
			break
		}
	}
	r = offset - cur
	return
}

func NewRequestGenerator(s [][]*Request, offset int) RequestGenerator {
	scripts := s
	sid, rid := getStartRequest(s, offset)
	count := s[sid][rid].Rate

	rg := func() *Request {
		nextId := func(s, r int) (int, int) {
			r++
			if r == len(scripts[s]) {
				r = 0
				s += 1
			}
			if s == len(scripts) {
				s = 0
			}
			if r < 0 {
				panic("wrong rid")
			}
			return s, r
		}

		if count == 0 {
			for {
				sid, rid = nextId(sid, rid)
				if scripts[sid][rid].Skip > 0 {
					scripts[sid][rid].Skip--
				} else {
					break
				}
			}
			count = scripts[sid][rid].Rate
		}
		count -= 1
		return scripts[sid][rid]
	}

	return rg
}
