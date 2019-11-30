package requests

import (
	"testing"
	"time"
)

func TestDefaultCollector(t *testing.T) {
	col := NewDefaultCollectot()

	col.Start()
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(5), 0})
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(10), 0})
	col.Add(&Meshure{nil, 200, time.Now(), time.Duration(15), 0})
	col.Stop()

	if col.Count != 3 {
		t.Errorf("Wrong number of inserted values: %d", col.Count)
	}
}
