package requests

import (
	"crypto/md5"
	"io"
	"time"
)

// Вынести в отдельный файл
type result struct {
	Error      error
	StatusCode int
	Start      time.Time
	Duration   time.Duration
	Sum        []byte
}

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
		w.Task.results <- &result{e, 0, s, t.Sub(s), nil}
		return e
	}

	if n, e := io.CopyN(mds, resp.Body, resp.ContentLength); e != nil || n != resp.ContentLength {
		w.Task.results <- &result{e, resp.StatusCode, s, t.Sub(s), nil}
		return e
	}

	w.Task.results <- &result{nil,
		resp.StatusCode,
		s,
		t.Sub(s),
		mds.Sum(nil)}
	return nil
}

func (w *worker) runWorkerLoop() {
	var throttle <-chan time.Time
	qps := w.Task.QPS

	if qps > 0 {
		throttle = time.Tick(time.Duration(1e6/(qps)) * time.Microsecond)
	}

	for i := 0; i < w.numRequests; i++ {
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
