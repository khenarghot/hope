package requests

import (
	"time"
)

type Meshure struct {
	Error      error
	StatusCode int
	Start      time.Time
	Duration   time.Duration
	Size       int64
}

type Collector interface {
	Add(m *Meshure)

	Start()
	Stop()

	Report() *Report
}

type DefaultCollector struct {
	queue chan *Meshure
	done  chan interface{}

	Count     int64
	Duration  time.Duration
	StartTime time.Time
	// Здесь будут поля для статистики, но потом
	TotalSize     int64
	TotalDuration time.Duration
	Slowest       time.Duration
	Fastest       time.Duration
	Codes         map[int]int
	Errors        map[string]int
}

func (cl *DefaultCollector) Report() *Report {
	if cl.Count == 0 {
		return nil
	}

	report := &Report{}
	report.Count = cl.Count

	report.DataSize = cl.TotalSize
	report.Duration = cl.Duration
	report.Start = cl.StartTime
	report.RawDuration = cl.TotalDuration
	if cl.Duration != time.Duration(0) {
		report.Rps = (int64(cl.Count) * int64(time.Second)) / int64(cl.Duration)
		report.AvgDuration = cl.TotalDuration / time.Duration(cl.Count)
	}

	if cl.Count != 0 {
		report.AvgSize = int64(cl.TotalSize) / int64(cl.Count)
	}
	report.Slowest = cl.Slowest
	report.Fastest = cl.Fastest
	if cl.Fastest > cl.Duration {
		report.Fastest = time.Duration(0)
	}

	ok := 0
	ko := 0
	for cd, cn := range cl.Codes {
		if cd > 199 && cd < 400 {
			ok += cn
		} else {
			ko += cn
		}
	}
	errs := 0
	for _, cnt := range cl.Errors {
		errs += cnt
	}

	report.OkResponse = ok
	report.OverResponse = ko
	report.Errors = errs
	report.NotFoundResponse = 0
	ff, have := cl.Codes[404]
	if have {
		report.NotFoundResponse = ff
	}

	return report
}

var BeginOfTimes time.Time = time.Unix(0, 0)

func NewDefaultCollectot() *DefaultCollector {
	c := &DefaultCollector{
		queue:     make(chan *Meshure),
		done:      make(chan interface{}),
		Count:     0,
		Duration:  time.Duration(0),
		StartTime: BeginOfTimes,
		Codes:     make(map[int]int, 1),
		Errors:    make(map[string]int),
		Fastest:   time.Hour,
	}

	return c
}

func (c *DefaultCollector) Add(m *Meshure) {
	c.queue <- m
}

func (c *DefaultCollector) Start() {
	go func() {
		for {
			m, ok := <-c.queue
			if ok {
				c.Count++
				if c.StartTime.Equal(BeginOfTimes) {
					c.StartTime = m.Start
				} else {
					c.Duration = time.Now().Sub(c.StartTime)
				}
				if m.Error != nil {
					c.Errors[m.Error.Error()]++
					continue
				}
				c.Codes[m.StatusCode]++
				c.TotalSize += m.Size
				c.TotalDuration += m.Duration
				if c.Slowest < m.Duration {
					c.Slowest = m.Duration
				}
				if c.Fastest > m.Duration {
					c.Fastest = m.Duration
				}
			} else {
				c.done <- nil
			}
		}
	}()
}

func (c *DefaultCollector) Stop() {
	close(c.queue)

	<-c.done
}
