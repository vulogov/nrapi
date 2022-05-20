[![Go](https://github.com/vulogov/nrapi/actions/workflows/go.yml/badge.svg)](https://github.com/vulogov/nrapi/actions/workflows/go.yml)

nrapi is a module providing access to a New Relic API for applications written in Golang.

## Pre-requirements.

New Relic is a SaaS Monitoring and telemetry provider and before you are going to use this module, you have to register with New Relic, get an account number and generate a license key. If you are not current user of New Relic, you can [establish free account](https://newrelic.com/signup) with New Relic and try this module. And [here](https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/), you can learn on how to create a new license key.

You have to set two environmental variables:

* NEWRELIC_ACCOUNT - this is you New Relic account number
* NEWRELIC_LICENSE_KEY - this is you New Relic license key.

## Installation

```shell
go get github.com/vulogov/nrapi
```

## Example on how to use nrapi module

This module will assist yo with sending of your custom metrics, events, application logs and traces to be processed and stored in New Relic SaaS service.

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()
```
This operation will create an instance that will be required for all further operations with New Relic API's

### How to send custom metric ?

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()

// nr.Gauge({HOSTNAME}, {TELEMETRY KEY}, {TELEMETRY VALUE})
nr.Gauge("test.example.com", "answer", 42)
```

### How to send a custom event ?

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()

// nr.Event({HOSTNAME}, {Destination}, {TELEMETRY KEY} {TELEMETRY VALUE})
nr.Event("test.example.com", "MYAPP", "answer", 42)
```

After event will be submitted to New Relic, you can query it using NRQL just like this

```
SELECT * FROM MYAPP
```

### How to send application logs to New Relic Log API ?

First, you have to create Log service and then using that service submit a log entries.

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()

// nr.LogService({LOG TYPE}, {SERVICE}, {HOST NAME})
log := nr.LogService("accesslog", "login-service", "test.example.com")
```
then you can use log.Send() to submit log entries

```Golang
log.Send("Login front end under a major attack #87962")
```

### How to instrument traces and segments ?

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()

// nr.Transaction({NAME OF TE TRANSACTION}, {HOSTNAME}, {SERVICE NAME})
txn := nr.Transaction("TXN1", "testhost.example.com", "IMPORTANTSERVICE")

// ... do something
txn.End()
```

If you transaction do have a multiple segments that you are looking to report separately, you can do it through Segment() call.

```Golang
import "github.com/vulogov/nrapi"

nr, err := nrapi.Create()

txn := nr.Transaction("TXN1", "testhost.example.com", "IMPORTANTSERVICE")

// ... do something
seg := txn.Segment("Segment name")

// ... do something
seg.End()

// ... do something
txn.End()
```
