package requests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSingleRequest(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("WWF\n"))
			}))
	defer ts.Close()

	task := NewTask(NewDefaultCollectot(), ts.Client().Transport, nil,
		0, 1, 10, time.Second)

	// Запускаем коллектор в ручную
	task.Collector.Start()

	wrk := &worker{
		Task:        task,
		numRequests: 10,
		nextRequest: nil,
	}

	req, e := NewRequest(ts.URL, "GET", nil, nil, 1)
	if e != nil {
		t.Errorf("Failed to create request")
	}

	if e := wrk.singleRequest(req); e != nil {
		t.Errorf("Filed to exexcute test request: %s", e.Error())
	}

	task.Collector.Stop()
}

func TestRunWorkerLoop(t *testing.T) {
	// Запускаем сервер
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Ok"))
			}))
	defer ts.Close()

	// Враппер для красивого вызова
	getRequest := func(url string, method string, header http.Header,
		body []byte, rate int) *Request {
		r, e := NewRequest(ts.URL+"/"+url, method, header, body, rate)
		if e != nil {
			panic(e.Error())
		}
		return r

	}

	clt := NewDefaultCollectot()
	requests := [][]*Request{{
		getRequest("/first", "GET", nil, nil, 1),
		getRequest("/second", "GET", nil, nil, 2)}, {
		getRequest("/一", "GET", nil, nil, 1),
		getRequest("/二", "GET", nil, nil, 2)}}
	task := NewTask(clt, ts.Client().Transport, requests,
		0, 2, 10, time.Second)

	// Запускаем коллектора. В норме через Init
	task.Collector.Start()

	// Создадим одного работника с пятью запросами
	wrk := &worker{
		Task:        task,
		numRequests: 5,
		nextRequest: NewRequestGenerator(task.Requests, 0),
	}

	wrk.runWorkerLoop()

	task.Collector.Stop()

	if clt.Count != 5 {
		t.Errorf("Wrong number of response: %d", clt.Count)
	}
}
