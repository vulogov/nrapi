package nrapi

import (
  "fmt"
  "bytes"
  "strings"
  "errors"
  "io/ioutil"
  "net/http"
  "context"
  "github.com/Jeffail/gabs/v2"
  "github.com/rocketlaunchr/dataframe-go"
  "github.com/rocketlaunchr/dataframe-go/imports"
  "github.com/vulogov/nrapi/mlog"
)

const NRQL_TPL = `
{
   actor {
      account(id: %v) {
         nrql(query: "%v") {
           results
         }
      }
   }
}
`

func DataFrame(res *gabs.Container) (*dataframe.DataFrame, error) {
  if res == nil {
    return nil, errors.New("Unknown dataset for loading ito DataFrame")
  }
  return imports.LoadFromJSON(context.TODO(), strings.NewReader(res.String()))
}

func (nr *NRAPI) NRQL(nrql string) *gabs.Container {
  pkt := gabs.New()
  query := fmt.Sprintf(NRQL_TPL, nr.NRAccount, nrql)
  pkt.Set(query, "query")
  mlog.Trace("QUERY: %v", pkt.String())
  res, err := NRGraphQLSendPayload(nr, []byte(pkt.String()))
  if err != nil {
    mlog.Trace("GraphQLAPI error: %v", err)
    return nil
  }
  out, err := gabs.ParseJSON([]byte(res))
  if err != nil {
    mlog.Trace("GraphQLAPI result parsing error: %v", err)
    return nil
  }
  result := out.Search("data", "actor", "account", "nrql", "results")
  if result != nil {
    return result
  }
  mlog.Trace("Missing results in NRQL response: %v", out.String())
  return nil
}

func NRGraphQLSendPayload(nr *NRAPI, payload []byte) (string, error) {
  req, err := http.NewRequest("POST", nr.GraphQLURL, bytes.NewBuffer(payload))
  if err != nil {
    mlog.Trace("GraphQLAPI error: %v", err)
    return "", err
  }
  req.Header.Set("Api-Key", nr.NRAPIKey)
  req.Header.Set("Content-Type", "application/json")
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    mlog.Trace("GraphQLAPI send error: %v", err)
    return "", err
  }
  defer resp.Body.Close()
  out, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    mlog.Trace("GraphQLAPI resp error: %v", err)
    return "", err
  }
  return string(out), nil
}
