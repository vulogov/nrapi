package nrapi

import (
	"time"
	"testing"
)

func TestNRAPI1(t *testing.T) {
	nr := New("", "")
	nr.MetricDelay = 1
	time.Sleep(2*time.Second)
	nr.Close()
}

func TestNRAPI2(t *testing.T) {
	nr, _ := Create()
	nr.MetricDelay = 1
	nr.Gauge("testhost", "answer", 42)
	time.Sleep(5*time.Second)
	nr.Close()
}
