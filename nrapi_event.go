package nrapi

import (
  "bytes"
  "time"
  "runtime"
  "io/ioutil"
  "compress/gzip"
  "net/http"
  "github.com/Jeffail/gabs/v2"
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

func NREventProcessor(nr *NRAPI) bool {
  runtime.Gosched()
  payload := gabs.New()
  payload.Array()

  for len(nr.EvtPipe) > 0 {
    pkt := <- nr.EvtPipe
    if pkt == nil {
      return false
    }
    mlog.Trace("Packet received(event): %v", pkt.String())
    payload.ArrayAppend(pkt)
  }
  NREventSendPayload(nr, payload)
  return true
}

func NREventSendPayload(nr *NRAPI, payload *gabs.Container) error {
  var gzbuf bytes.Buffer

  w := gzip.NewWriter(&gzbuf)
  w.Write([]byte(payload.String()))
  w.Close()
  gzpayload := []byte(gzbuf.Bytes())
  req, err := http.NewRequest("POST", nr.EventURL, bytes.NewBuffer(gzpayload))
  if err != nil {
    mlog.Trace("EventAPI error: %v", err)
    return err
  }
  req.Header.Set("Api-Key", nr.NRLicenseKey)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Content-Encoding", "gzip")
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    mlog.Trace("EventAPI send error: %v", err)
    return err
  }
  defer resp.Body.Close()
  out, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    mlog.Trace("EventAPI resp error: %v", err)
    return err
  }
  mlog.Trace(string(out))
  return nil
}

func (nr *NRAPI) Event(host string, name string, key string, value interface{}, attributes ...map[string]interface{}) error {
  pkt := gabs.New()

  pkt.Set(name, "eventType")
  pkt.Set(host, "hostname")
  pkt.Set(time.Now().UnixNano() / int64(time.Millisecond), "timestamp")
  if len(nr.ApplicationID) > 0 {
    pkt.Set(nr.ApplicationID, "applicationID")
  }
  if len(nr.ApplicationName) > 0 {
    pkt.Set(nr.ApplicationName, "applicationName")
  }
  for _, d := range(attributes) {
    for k, v := range(d) {
      pkt.Set(v, k)
    }
  }
  pkt.Set(value, key)
  nr.EvtPipe <- pkt
  return nil
}
