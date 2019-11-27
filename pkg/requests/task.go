package requests

import (
	"crypto/tls"
	op "github.com/khenarghot/hope/pkg/options"
	"golang.org/x/net/http2"
	"net/http"
	"sync"
	"time"
)

// Интерфейс для реализации запуска зададчи
type Work interface {
	Run()
	Stop()
	Finish()
}

// Task описание задаваемой нагрузки получаемый из конфигурации
type Task struct {
	*http.Client
	// unused: fix it
	*http.Transport
	H2          bool
	Requests    [][]*Request
	QPS         float64
	Workers     int
	NumRequests int
	Host        string
	Timeout     time.Duration

	results chan *result
	stop    chan interface{}
}

func (t *Task) Stop() {
	for i := 0; i < t.Workers; i++ {
		t.stop <- nil
	}
}

func (t *Task) runWorks() {
	var wg sync.WaitGroup
	wg.Add(t.Workers)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         t.Host,
		},
		MaxIdleConnsPerHost: min(t.Workers, op.HopeConfig.Core.Connections),
		DisableCompression:  true,
		DisableKeepAlives:   true,
		Proxy:               nil,
	}

	if t.H2 {
		http2.ConfigureTransport(tr)
	} else {
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}
	t.Client = &http.Client{
		Transport: tr,
		Timeout:   t.Timeout,
	}

	t.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	var i int
	whole := t.NumRequests / t.Workers
	for i = 0; i < t.Workers-1; i++ {
		wrk := &worker{t, whole, NewRequestGenerator(t.Requests)}
		go func() {
			wrk.runWorkerLoop()
			wg.Done()
		}()
	}
	wrk := &worker{t, whole + t.NumRequests%t.Workers, NewRequestGenerator(t.Requests)}
	go func() {
		wrk.runWorkerLoop()
		wg.Done()
	}()
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
