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

type NRLOGAPI struct {
  Nr          *NRAPI
  Header      *gabs.Container
}

func NRLogDaemon(nr *NRAPI) {
  nr.Wg.Add(1)
  defer nr.Wg.Done()
  c := 0
  mlog.Trace("LOG daemon entering the loop")
  out:
  for {
    if len(nr.LogPipe) > 0 {
      if c > nr.LogDelay {
        mlog.Trace("Running LOG processor")
        if ! NRLogProcessor(nr) {
          break out
        }
        c = 0
      }
    }
    time.Sleep(1 *time.Second)
    c += 1
  }
  mlog.Trace("LOG daemon exiting the loop")
}

func NRLogProcessor(nr *NRAPI) bool {
  runtime.Gosched()
  payload := gabs.New()
  payload.Array()
  l := gabs.New()
  payload.ArrayAppend(l)
  l.Array("logs")
  for len(nr.LogPipe) > 0 {
    pkt := <- nr.LogPipe
    if pkt == nil {
      return false
    }
    mlog.Trace("Packet received(log): %v", pkt.String())
    l.ArrayAppend(pkt, "logs")
  }
  NRLogSendPayload(nr, payload)
  return true
}

func NRLogSendPayload(nr *NRAPI, payload *gabs.Container) error {
  var gzbuf bytes.Buffer

  w := gzip.NewWriter(&gzbuf)
  w.Write([]byte(payload.String()))
  w.Close()
  gzpayload := []byte(gzbuf.Bytes())
  req, err := http.NewRequest("POST", nr.LogURL, bytes.NewBuffer(gzpayload))
  if err != nil {
    mlog.Trace("LogAPI error: %v", err)
    return err
  }
  req.Header.Set("Api-Key", nr.NRLicenseKey)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Content-Encoding", "gzip")
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    mlog.Trace("LogAPI send error: %v", err)
    return err
  }
  defer resp.Body.Close()
  out, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    mlog.Trace("LogAPI resp error: %v", err)
    return err
  }
  mlog.Trace(string(out))
  return nil
}

func (nr *NRAPI) Log(attributes ...map[string]interface{}) *NRLOGAPI {
  res := new(NRLOGAPI)
  res.Nr = nr
  res.Header = gabs.New()
  for _, d := range(attributes) {
    for k, v := range(d) {
      res.Header.Set(v, "attributes", k)
    }
  }
  return res
}

func (nr *NRAPI) LogService(logtype string, service string, hostname string, attributes ...map[string]interface{}) *NRLOGAPI {
  res := new(NRLOGAPI)
  res.Nr = nr
  res.Header = gabs.New()
  for _, d := range(attributes) {
    for k, v := range(d) {
      res.Header.Set(v, "attributes", k)
    }
  }
  res.Header.Set(logtype, "attributes", "logtype")
  res.Header.Set(service, "attributes", "service")
  res.Header.Set(hostname, "attributes", "hostname")
  return res
}

func (nrl *NRLOGAPI) Send(msg string, attributes ...map[string]interface{}) error {
  pkt := gabs.New()
  pkt.Merge(nrl.Header)
  pkt.Set(time.Now().UnixNano() / int64(time.Millisecond), "timestamp")
  pkt.Set(msg, "message")
  nrl.Nr.LogPipe <- pkt
  return nil
}
