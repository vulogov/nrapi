package nrapi

import (
  "bytes"
  "time"
  "runtime"
  "io/ioutil"
  "compress/gzip"
  "net/http"
  "github.com/google/uuid"
  "github.com/Jeffail/gabs/v2"
  "github.com/vulogov/nrapi/mlog"
)

type NRTRANSACTION struct {
  Nr          *NRAPI
  Name        string
  Id          string
  TraceId     string
  Host        string
  ServiceName string
  Timestamp   int64
  ErrorMsg    string
  Ended       bool
  Header      *gabs.Container
}

type NRSEGMENT struct {
  Txn         *NRTRANSACTION
  Name        string
  Id          string
  Timestamp   int64
  ErrorMsg    string
  Ended       bool
  Header      *gabs.Container
}

func (nr *NRAPI) Transaction(name string, host string, service string, attributes ...map[string]interface{}) *NRTRANSACTION {
  txn := new(NRTRANSACTION)
  txn.Id        = uuid.NewString()
  txn.Nr        = nr
  txn.TraceId   = uuid.NewString()
  txn.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
  txn.Name      = name
  txn.Host      = host
  txn.ServiceName = service
  txn.Ended       = false
  txn.Header      = gabs.New()
  for _, d := range(attributes) {
    for k, v := range(d) {
      txn.Header.Set(v, "attributes", k)
    }
  }
  return txn
}

func (txn *NRTRANSACTION) SetError(msg string) {
  txn.ErrorMsg = msg
}

func (txn *NRTRANSACTION) End() {
  stamp := time.Now().UnixNano() / int64(time.Millisecond)
  pkt := gabs.New()
  pkt.Merge(txn.Header)
  pkt.Set(txn.Id, "id")
  pkt.Set(txn.TraceId, "trace.id")
  pkt.Set(txn.Name, "attributes", "name")
  pkt.Set(txn.Host, "attributes", "host")
  pkt.Set(txn.ServiceName, "attributes", "service.name")
  pkt.Set(stamp-txn.Timestamp, "attributes", "duration.ms")
  if len(txn.ErrorMsg) > 0 {
    pkt.Set(txn.ErrorMsg, "attributes", "error.message")
  }
  txn.Ended = true
  txn.Nr.TracePipe <- pkt
}

func (txn *NRTRANSACTION) Segment(name string, attributes ...map[string]interface{}) *NRSEGMENT {
  if txn.Ended {
    mlog.Trace("Attempt to create segment for finished transaction")
    return nil
  }
  seg := new(NRSEGMENT)
  seg.Id        = uuid.NewString()
  seg.Name      = name
  seg.Txn       = txn
  seg.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
  seg.Ended     = false
  seg.Header    = gabs.New()
  for _, d := range(attributes) {
    for k, v := range(d) {
      seg.Header.Set(v, "attributes", k)
    }
  }
  return seg
}

func (seg *NRSEGMENT) SetError(msg string) {
  seg.ErrorMsg = msg
}

func (seg *NRSEGMENT) End() {
  stamp := time.Now().UnixNano() / int64(time.Millisecond)
  pkt := gabs.New()
  pkt.Merge(seg.Header)
  pkt.Merge(seg.Txn.Header)
  pkt.Set(seg.Id, "id")
  pkt.Set(seg.Txn.TraceId, "trace.id")
  pkt.Set(seg.Name, "attributes", "name")
  pkt.Set(seg.Txn.Host, "attributes", "host")
  pkt.Set(seg.Txn.ServiceName, "attributes", "service.name")
  pkt.Set(stamp-seg.Timestamp, "attributes", "duration.ms")
  if len(seg.ErrorMsg) > 0 {
    pkt.Set(seg.ErrorMsg, "attributes", "error.message")
  }
  seg.Ended = true
  seg.Txn.Nr.TracePipe <- pkt
}

func NRTraceDaemon(nr *NRAPI) {
  nr.Wg.Add(1)
  defer nr.Wg.Done()
  c := 0
  mlog.Trace("TRACE daemon entering the loop")
  out:
  for {
    if len(nr.TracePipe) > 0 {
      if c > nr.TraceDelay {
        mlog.Trace("Running TRACE processor")
        if ! NRTraceProcessor(nr) {
          break out
        }
        c = 0
      }
    }
    time.Sleep(1 *time.Second)
    c += 1
  }
  mlog.Trace("TRACE daemon exiting the loop")
}

func NRTraceProcessor(nr *NRAPI) bool {
  runtime.Gosched()
  pkt := gabs.New()
  pkt.Array()
  s := gabs.New()
  pkt.ArrayAppend(s)
  s.Array("spans")

  for len(nr.TracePipe) > 0 {
    pkt := <- nr.TracePipe
    if pkt == nil {
      return false
    }
    mlog.Trace("Packet received(trace): %v", pkt.String())
    s.ArrayAppend(pkt, "spans")
  }
  NRTraceSendPayload(nr, pkt)
  return true
}

func NRTraceSendPayload(nr *NRAPI, payload *gabs.Container) error {
  var gzbuf bytes.Buffer

  mlog.Trace("TRACE: %v", payload.String())
  w := gzip.NewWriter(&gzbuf)
  w.Write([]byte(payload.String()))
  w.Close()
  gzpayload := []byte(gzbuf.Bytes())
  req, err := http.NewRequest("POST", nr.TraceURL, bytes.NewBuffer(gzpayload))
  if err != nil {
    mlog.Trace("TraceAPI error: %v", err)
    return err
  }
  req.Header.Set("Api-Key", nr.NRLicenseKey)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Content-Encoding", "gzip")
  req.Header.Set("Data-Format", "newrelic")
  req.Header.Set("Data-Format-Version", "1")


  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    mlog.Trace("TraceAPI send error: %v", err)
    return err
  }
  defer resp.Body.Close()
  out, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    mlog.Trace("TraceAPI resp error: %v", err)
    return err
  }
  mlog.Trace(string(out))
  return nil
}
