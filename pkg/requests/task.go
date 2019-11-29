package requests

import (
	"net/http"
	"sync"
	"time"
)

// Интерфейс для реализации запуска зададчи
type Work interface {
	Run()
	Init()
	Stop()
}

// Task описание задаваемой нагрузки получаемый из конфигурации
type Task struct {
	Collector
	*http.Client
	Requests    [][]*Request
	QPS         float64
	Workers     int
	NumRequests int

	started bool
	stop    chan interface{}
	wg      sync.WaitGroup
}

func NewTask(c Collector, rt http.RoundTripper, requests [][]*Request,
	qps float64, w, r int, timeout time.Duration) *Task {
	ts := &Task{
		Collector: c,
		Client: &http.Client{
			Transport: rt,
			Timeout:   timeout,
		},
		Requests:    requests,
		QPS:         qps,
		Workers:     w,
		NumRequests: r,
		started:     false,
		stop:        make(chan interface{}),
	}
	ts.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return ts
}

func (t *Task) Init() {
	t.Collector.Start()
}

func (t *Task) Stop() {
	if !t.started {
		return
	}

	for i := 0; i < t.Workers; i++ {
		t.stop <- nil
	}
	t.wg.Wait()
	t.Collector.Stop()
}

func (t *Task) Run() {
	t.wg.Add(t.Workers)
	t.started = true

	var i int
	whole := t.NumRequests / t.Workers
	for i = 0; i < t.Workers-1; i++ {
		wrk := &worker{t, whole, NewRequestGenerator(t.Requests)}
		go func() {
			wrk.runWorkerLoop()
			t.wg.Done()
		}()
	}
	wrk := &worker{t, whole + t.NumRequests%t.Workers, NewRequestGenerator(t.Requests)}
	go func() {
		wrk.runWorkerLoop()
		t.wg.Done()
	}()
	t.wg.Wait()
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
