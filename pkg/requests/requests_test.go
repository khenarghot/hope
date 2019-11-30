package requests

import (
	"net/http"
	"testing"
)

func TestNewRequestsGenerator(t *testing.T) {
	sc := [][]*Request{{
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}, {
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}}
	gen := NewRequestGenerator(sc, 0)

	first := gen()
	count := 0

	for count < 10 {
		count += 1
		cur := gen()
		if cur == first {
			break
		}
	}

	if count != 6 {
		t.Errorf("Wrong number of iterations %d", count)
	}
}

func TestNewRequestsGenerator100(t *testing.T) {
	sc := [][]*Request{{
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}, {
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}}
	gen := NewRequestGenerator(sc, 100)

	first := gen()
	count := 0

	for count < 10 {
		count += 1
		cur := gen()
		if cur == first {
			break
		}
	}

	if count != 6 {
		t.Errorf("Wrong number of iterations %d", count)
	}
}

func TestNewRequestsGenerator3(t *testing.T) {
	sc := [][]*Request{{
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}, {
		&Request{http.Request{}, nil, nil, 1, 0},
		&Request{http.Request{}, nil, nil, 2, 0}}}
	gen := NewRequestGenerator(sc, 0)

	first := gen()
	count := 0

	for count < 10 {
		count += 1
		cur := gen()
		if cur == first {
			break
		}
	}

	if count != 6 {
		t.Errorf("Wrong number of iterations %d", count)
	}
}
