package requests

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
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
	respHash, _ := hex.DecodeString("8efb81886a9b466e3913dedbc1c79fb0")

	up, e := url.Parse(ts.URL)
	if e != nil {
		t.Errorf("Parsing test error")
	}

	task := &Task{
		Client:      ts.Client(),
		Transport:   nil,
		Requests:    nil,
		QPS:         0,
		Workers:     1,
		NumRequests: 10,
		Host:        up.Host,
		Timeout:     time.Second,

		results: make(chan *result),
		stop:    make(chan interface{}),
	}

	wrk := &worker{
		Task:        task,
		numRequests: 10,
		nextRequest: nil,
	}

	donech := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case r := <-task.results:
				if bytes.Compare(r.Sum, respHash) != 0 {
					t.Errorf("Wrong response hash")
				}
			case <-donech:
				return
			}
		}
	}()

	req, e := NewRequest(ts.URL, "GET", nil, nil, 1)
	if e != nil {
		t.Errorf("Failed to create request")
	}

	if e := wrk.singleRequest(req); e != nil {
		t.Errorf("Filed to exexcute test request: %s", e.Error())
	}
	donech <- nil
	wg.Wait()
}

func TestRunWorkerLoop(t *testing.T) {
	// Запускаем сервер
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Ok"))
			}))
	defer ts.Close()

	// Для получение порта и адреса
	up, e := url.Parse(ts.URL)
	if e != nil {
		t.Errorf("Parsing test error")
	}

	// Враппер для красивого вызова
	getRequest := func(url string, method string, header http.Header,
		body []byte, rate int) *Request {
		r, e := NewRequest(ts.URL+"/"+url, method, header, body, rate)
		if e != nil {
			panic(e.Error())
		}
		return r

	}

	// Вручную создаём  задачу
	task := &Task{
		Client:    ts.Client(),
		Transport: nil,
		Requests: [][]*Request{{
			getRequest("/first", "GET", nil, nil, 1),
			getRequest("/second", "GET", nil, nil, 2)}, {
			getRequest("/一", "GET", nil, nil, 1),
			getRequest("/二", "GET", nil, nil, 2)}},
		QPS:         0,
		Workers:     2,
		NumRequests: 10,
		Host:        up.Host,
		Timeout:     time.Second,

		results: make(chan *result),
		stop:    make(chan interface{}),
	}

	// Создадим сборщик статистики
	var responseCount int = 0
	var wg sync.WaitGroup
	wg.Add(1)
	doneChan := make(chan interface{})
	// Функция сбора ответов
	go func() {
		defer wg.Done()
		for {
			select {
			case <-task.results:
				responseCount += 1
			case <-doneChan:
				return
			}
		}
	}()

	// Создадим одного работника с пятью запросами
	wrk := &worker{
		Task:        task,
		numRequests: 5,
		nextRequest: NewRequestGenerator(task.Requests),
	}

	wrk.runWorkerLoop()

	doneChan <- nil
	wg.Wait()

	if responseCount != 5 {
		t.Errorf("Wrong number of response: %d", responseCount)
	}
}
