package requests

import (
	"testing"
	"time"
)

func TestDefaultCollector(t *testing.T) {
	col := NewDefaultCollectot()

	col.Start()
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(5), nil})
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(10), nil})
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(15), nil})
	col.Stop()

	if col.Count != 3 {
		t.Errorf("Wrong number of inserted values: %d", col.Count)
	}
}
