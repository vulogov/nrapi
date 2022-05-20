package nrapi

import (
  "fmt"
  "sync"
  "os"
  "errors"
  "github.com/vulogov/nrapi/mlog"
  "github.com/Jeffail/gabs/v2"
)

const MAXPIPECAP = 1000000

type NRAPI struct {
  NRLicenseKey        string
  NRAccount           string
  ApplicationID       string
  ApplicationName     string
  MetricURL           string
  EventURL            string
  LogURL              string
  TraceURL            string
  Logfilename         string
  EvtPipe             chan *gabs.Container
  MetricPipe          chan *gabs.Container
  TracePipe           chan *gabs.Container
  LogPipe             chan *gabs.Container
  Wg                  sync.WaitGroup
  MetricDelay         int
  EventDelay          int
  LogDelay            int
  TraceDelay          int
  CanSubmit           bool
}

func Create() (*NRAPI, error) {
  var account, license string
  var ok bool

  if account, ok = os.LookupEnv("NEWRELIC_ACCOUNT"); ! ok {
    return nil, errors.New("NEWRELIC_ACCOUNT environment variable is not set")
  }
  if license, ok = os.LookupEnv("NEWRELIC_LICENSE_KEY"); ! ok {
    return nil, errors.New("NEWRELIC_LICENSE_KEY environment variable is not set")
  }
  return New(account, license), nil
}

func New(account string, license_key string) *NRAPI {
  nr := new(NRAPI)
  nr.NRAccount     = account
  nr.NRLicenseKey  = license_key
  nr.EvtPipe       = make(chan *gabs.Container, MAXPIPECAP)
  nr.MetricPipe    = make(chan *gabs.Container, MAXPIPECAP)
  nr.TracePipe     = make(chan *gabs.Container, MAXPIPECAP)
  nr.LogPipe       = make(chan *gabs.Container, MAXPIPECAP)
  nr.CanSubmit     = true
  nr.MetricDelay   = 60
  nr.EventDelay    = 5
  nr.LogDelay      = 5
  nr.TraceDelay    = 5
  nr.MetricURL     = "https://metric-api.newrelic.com/metric/v1"
  nr.EventURL      = fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%v/events", account)
  nr.LogURL        = "https://log-api.newrelic.com/log/v1"
  nr.TraceURL      = "https://trace-api.newrelic.com/trace/v1"
  nr.SetConsoleLog()
  go NRMetricDaemon(nr)
  go NREventDaemon(nr)
  go NRLogDaemon(nr)
  go NRTraceDaemon(nr)
  nr.DisableLog()
  return nr
}

func (nr *NRAPI) SetApplication(id string, name string) {
  nr.ApplicationID = id
  nr.ApplicationName = name
}

func (nr *NRAPI) SetConsoleLog() {
  mlog.Stop()
  mlog.Start(mlog.LevelTrace, "")
}

func (nr *NRAPI) DisableLog() {
  mlog.Stop()
}

func (nr *NRAPI) SetLoggingToFile(name string) {
  mlog.Stop()
  mlog.Start(mlog.LevelTrace, name)
}

func (nr *NRAPI) SetLoggingToDefaultFile() {
  mlog.Stop()
  mlog.Start(mlog.LevelTrace, fmt.Sprintf("nrapi-%v.log", nr.NRAccount))
}

func (nr *NRAPI) Close() {
  nr.CanSubmit   = false
  nr.EvtPipe     <- nil
  nr.MetricPipe  <- nil
  nr.TracePipe   <- nil
  nr.LogPipe     <- nil
  nr.Wg.Wait()
  mlog.Stop()
}
