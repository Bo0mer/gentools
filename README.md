# gentools

Gentools provides a collection of tools that help by generating Go interface
implementations that add monitoring, logging and tracing.

## Installation

In order to install all tools run:
```
go get -u github.com/Bo0mer/gentools/cmd/...
```

## Using mongen

Given a path to a package and an interface name, you could generate monitoring
implementation of the interface.

```go
package service

import "context"

type Service interface {
    DoWork(context.Context, int, string) (string, error)
}
```

```bash
$ mongen path/to/service Service
Wrote monitoring implementation of "path/to/service.Service" to "path/to/service/servicews/monitoring_service.go"
```

### Using monitoring implementation in your program

Instantiate monitoring implementations with `NewMonitoring{InterfaceName}`:

```go
var svc Service = service.New()
svc = servicemws.NewMonitoringService(svc, totalOps, faildOps, opsDuration)
```

## Using logen

Given a path to a package, an interface name and logger _(optional)_, you could generate logging
implementation of the interface. 

### Supported loggers
* `go_kit_log` - The logger provided from [go-kit](https://github.com/go-kit/kit)
* `logrus`  - The logger provided from [logrus](https://github.com/sirupsen/logrus)
* `stdlog`  - The logger provided from the standard go library logger

If logger is not mentioned, the app will generate its implementation by using the standard go library logger. 

At the moment, only error logging is supported.

## Using tracegen

Given a path to a package and an interface name, you could generate tracing
implementation of the interface. Tracing will be added only to methods that
take a `context.Context` as a first argument. All other methods will be proxied
to the original implementation, without any modifications or additions.

## Integration with go generate

The best way to integrate the tools within your project is to use the
`go:generate` directive in your code.

```bash
$ cat path/to/service/file.go
```

```go
package service

import "context"

//go:generate mongen . Service
//go:generate tracegen . Service
//go:generate logen . Service stdlog

type Service interface {
	DoWork(context.Context, int, string) (string, error)
}
```

```
$ go generate ./...
Wrote monitoring implementation of "path/to/service.Service" to "servicemws/monitoring_service.go"
Wrote tracing implementation of "path/to/service.Service" to "servicemws/tracing_service.go"
Wrote logging implementation of "path/to/service.Service" to "servicemws/logging_service.go"
```

## Credits

* Special thanks to [Momchil Atanasov](https://github.com/mokiat) and his
  awesome project [gostub](https://github.com/mokiat/gostub), which code was
  used as a starting point for developing gentools v2.
