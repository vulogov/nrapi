package nrapi

import (
  "time"
  "runtime"
  "github.com/vulogov/nrapi/mlog"
)

func NREventDaemon(nr *NRAPI) {
  nr.Wg.Add(1)
  defer nr.Wg.Done()
  c := 0
  mlog.Trace("EVENT daemon entering the loop")
  out:
  for {
    if len(nr.EvtPipe) > 0 {
      if c > nr.EventDelay {
        mlog.Trace("Running EVENT processor")
        if ! NREventProcessor(nr) {
          break out
        }
        c = 0
      }
    }
    time.Sleep(1 *time.Second)
    c += 1
  }
  mlog.Trace("EVENT daemon exiting the loop")
}

func NREVENTProcessor(nr *NRAPI) bool {
  runtime.Gosched()
  pkt := gabs.New()
  pkt.Array()

  for len(nr.EvtPipe) > 0 {
    pkt := <- nr.EvtPipe
    if pkt == nil {
      return false
    }
    mlog.Trace("Packet received(event): %v", pkt.String())
    m.ArrayAppend(pkt)
  }
  NREventSendPayload(nr, pkt)
  return true
}

func NREventSendPayload(nr *NRAPI, payload *gabs.Container) error {
  var gzbuf bytes.Buffer

  w := gzip.NewWriter(&gzbuf)
  w.Write([]byte(payload.String()))
  w.Close()
  gzpayload := []byte(gzbuf.Bytes())
  req, err := http.NewRequest("POST", nr.MetricURL, bytes.NewBuffer(gzpayload))
  if err != nil {
    mlog.Trace("MetricAPI error: %v", err)
    return err
  }
  req.Header.Set("X-Insert-Key", nr.NRLicenseKey)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Content-Encoding", "gzip")
  client := &http.Client{}
  resp, err := client.Do(req)
  defer resp.Body.Close()
  if err != nil {
    mlog.Trace("MetricAPI send error: %v", err)
    return err
  }
  out, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    mlog.Trace("MetricAPI resp error: %v", err)
    return err
  }
  mlog.Trace(string(out))
  return nil
}

func (nr *NRAPI) Metric(host string, name string, value interface{}, mt string, attributes ...map[string]interface{}) error {
  pkt := gabs.New()
  switch v := value.(type) {
  case float64:
    pkt.Set(v, "value")
  case int64:
    pkt.Set(float64(v), "value")
  case int32:
    pkt.Set(float64(v), "value")
  case int:
    pkt.Set(float64(v), "value")
  case float32:
    pkt.Set(float64(v), "value")
  case string:
    val, err := strconv.ParseFloat(v, 64)
    if err != nil {
      return err
    }
    pkt.Set(val, "value")
  }
  pkt.Set(name, "name")
  pkt.Set(mt, "type")
  pkt.Set(host, "attributes", "hostname")
  pkt.Set(time.Now().UnixNano() / int64(time.Millisecond), "timestamp")
  if len(nr.ApplicationID) > 0 {
    pkt.Set(nr.ApplicationID, "attributes", "applicationID")
  }
  if len(nr.ApplicationName) > 0 {
    pkt.Set(nr.ApplicationName, "attributes", "applicationName")
  }
  for _, d := range(attributes) {
    for k, v := range(d) {
      pkt.Set(v, "attributes", k)
    }
  }
  nr.MetricPipe <- pkt
  return nil
}

func (nr *NRAPI) Gauge(host string, name string, value interface{}, attributes ...map[string]interface{}) error {
  return nr.Metric(host, name, value, "gauge", attributes...)
}
