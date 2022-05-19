package nrapi

import (
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
  EvtPipe             chan *gabs.Container
  MetricPipe          chan *gabs.Container
  TracePipe           chan *gabs.Container
  LogPipe             chan *gabs.Container
  Wg                  sync.WaitGroup
  MetricDelay         int
  EventDelay          int
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
  nr.MetricURL     = "https://metric-api.newrelic.com/metric/v1"
  nr.EventURL      = fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%v/events", account)
  mlog.Start(mlog.LevelTrace, "")
  mlog.Info("NRAPI instance created")
  go NRMetricDaemon(nr)
  go NREventDaemon(nr)
  return nr
}

func (nr *NRAPI) SetApplication(id string, name string) {
  nr.ApplicationID = id
  nr.ApplicationName = name
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
