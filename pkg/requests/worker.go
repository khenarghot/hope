package requests

import (
	"crypto/md5"
	"io"
	"time"
)

type worker struct {
	*Task
	numRequests int

	nextRequest RequestGenerator
}

func (w *worker) singleRequest(r *Request) error {
	// TODO: Добавить нормальнльную статистику
	s := time.Now()
	resp, e := w.Task.Client.Do(&r.Request)
	t := time.Now()
	mds := md5.New()

	if e != nil {
		w.Task.Collector.Add(&Meshure{e, 0, s, t.Sub(s), 0, nil})
		return e
	}

	if n, e := io.CopyN(mds, resp.Body, resp.ContentLength); e != nil || n != resp.ContentLength {
		w.Task.Collector.Add(&Meshure{e, resp.StatusCode, s, t.Sub(s), resp.ContentLength, nil})
		return e
	}

	w.Task.Collector.Add(&Meshure{nil,
		resp.StatusCode,
		s,
		t.Sub(s),
		resp.ContentLength,
		mds.Sum(nil)})
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
