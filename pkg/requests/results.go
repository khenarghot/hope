package requests

import (
	"time"
)

type Meshure struct {
	Error      error
	StatusCode int
	Start      time.Time
	Duration   time.Duration
	Sum        []byte
}

type Collector interface {
	Add(m *Meshure)

	Start()
	Stop()
}

type DefaultCollector struct {
	queue chan *Meshure
	done  chan interface{}

	Count     int64
	Duration  time.Duration
	StartTime time.Time
	// Здесь будут поля для статистики, но потом
}

var BeginOfTimes time.Time = time.Unix(0, 0)

func NewDefaultCollectot() *DefaultCollector {
	c := &DefaultCollector{
		queue:     make(chan *Meshure),
		done:      make(chan interface{}),
		Count:     0,
		Duration:  time.Duration(0),
		StartTime: BeginOfTimes,
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
