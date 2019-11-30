package requests

import (
	"io"
	"io/ioutil"
	"time"
)

type worker struct {
	*Task
	numRequests int

	nextRequest RequestGenerator
}

func (w *worker) singleRequest(r *Request) error {
	s := time.Now()
	resp, e := w.Task.Client.Do(&r.Request)
	t := time.Now()

	if e != nil {
		w.Task.Collector.Add(&Meshure{e, 0, s, t.Sub(s), 0})
		return e
	}

	if n, e := io.Copy(ioutil.Discard, resp.Body); e != nil || n != resp.ContentLength {
		// Если у нас сбоит получение тела репортим об ошибке
		w.Task.Collector.Add(&Meshure{nil, resp.StatusCode, s, t.Sub(s), n})
		resp.Body.Close()
		return e
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		// Если у нас нет данных мы не будем ходить сюда некоторое число итераций
		r.Skip = resp.StatusCode
	}

	w.Task.Collector.Add(&Meshure{nil,
		resp.StatusCode,
		s,
		t.Sub(s),
		resp.ContentLength,
	})
	return nil
}

func (w *worker) runWorkerLoop() {
	var throttle <-chan time.Time
	qps := w.Task.QPS

	if qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(qps)) * time.Microsecond)
	}

	for i := 0; w.numRequests < 0 || i < w.numRequests; i++ {
		select {
		case <-w.Task.stop:
			return
		default:
			if qps > 0 {
				<-throttle
			}
			req := w.nextRequest()
			w.singleRequest(req)
		}
	}
}
