package nrapi

import (
	"time"
	"testing"
)

func TestNRAPI1(t *testing.T) {
	nr := New("", "")
	nr.EventDelay = 1
	nr.MetricDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	time.Sleep(2*time.Second)
	nr.Close()
}

func TestNRAPI2(t *testing.T) {
	nr, _ := Create()
	nr.SetConsoleLog()
	nr.MetricDelay = 1
	nr.EventDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	nr.Gauge("testhost", "answer", 42)
	time.Sleep(5*time.Second)
	nr.Close()
}

func TestNRAPI3(t *testing.T) {
	nr, _ := Create()
	nr.SetConsoleLog()
	nr.EventDelay = 1
	nr.MetricDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	nr.Event("testhost", "NRAPI", "answer", 42)
	time.Sleep(5*time.Second)
	nr.Close()
}

func TestNRAPI4(t *testing.T) {
	nr, _ := Create()
	nr.SetConsoleLog()
	nr.EventDelay = 1
	nr.MetricDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	log := nr.LogService("accesslog", "login-service", "testhost.example.com")
	log.Send("This is a test message")
	time.Sleep(5*time.Second)
	nr.Close()
}

func TestNRAPI5(t *testing.T) {
	nr, _ := Create()
	nr.SetConsoleLog()
	nr.EventDelay = 1
	nr.MetricDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	txn := nr.Transaction("TXN1", "testhost.example.com", "IMPORTANTSERVICE")
	seg := txn.Segment("BITTY SEGMENT")
	time.Sleep(1*time.Second)
	seg.End()
	time.Sleep(1*time.Second)
	txn.End()
	time.Sleep(5*time.Second)
	nr.Close()
}

func TestNRAPI6(t *testing.T) {
	nr, _ := Create()
	nr.SetConsoleLog()
	nr.EventDelay = 1
	nr.MetricDelay = 1
	nr.LogDelay = 1
	nr.TraceDelay = 1
	txn := nr.Transaction("TXN1", "testhost.example.com", "IMPORTANTSERVICE")
	txn.SetError("Oops, error happens")
	time.Sleep(1*time.Second)
	txn.End()
	time.Sleep(5*time.Second)
	nr.Close()
}
