package requests


import (
	"testing"
	"net/http"
)

func TestNewRequestsGenerator(t *testing.T) {
	sc := [][]*Request {{
		&Request{http.Request{}, nil, false, nil, 1},
		&Request{http.Request{}, nil, false, nil, 2}},{
		&Request{http.Request{}, nil, false, nil, 1},
		&Request{http.Request{}, nil, false, nil, 2}}}			
	gen := NewRequestGenerator(sc)

	first := gen()
	count := 0

	for count < 10 {
		count += 1 
		cur := gen()
		if (cur == first) {
			break
		}
	}

	if count != 6 {
		t.Errorf("Wrong number of iterations %d", count)
	}
}
			
			
