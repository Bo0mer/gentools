# gentools

Gentools provides a collection of tools that help by generating Go interface
implementations that add logging or tracing. 

## Using logen

Command logen generates interface implementations that add logging. There are
two options for the log level:
* `error` - log only invocations that result in a non-nil error.
* `debug` - log all invocations with input & output parameters, method name and
  invocation duration. 

Let's see an examle. Given the following Go interface definition:

```go
package example // import "github.com/Bo0mer/gentools/pkg/example"

type Interface interface {
	Foo(ctx context.Context, a, b int) (err error)
	Bar(ctx context.Context, c string) (x string, y int, err error)
}
```

`logen` needs the following invocation:

`$ logen github.com/Bo0mer/gentools/pkg/example Interface error | gofmt`

and will produce the following code:

```go
package example

import (
	"context"

	"github.com/go-kit/kit/log"
)

// NewErrorLoggingInterface logs all non-nil errors.
func NewErrorLoggingInterface(next Interface, log log.Logger) Interface {
	return &errorLoggingInterface{
		next: next,
		log:  log,
	}
}

type errorLoggingInterface struct {
	log  log.Logger
	next Interface
}

func (l *errorLoggingInterface) Bar(ctx context.Context, c string) (x string, y int, err error) {
	x, y, err = l.next.Bar(ctx, c)

	if err != nil {
		l.log.Log(
			"method", "Bar",
			"error", err.Error(),
		)
	}
	return x, y, err
}

func (l *errorLoggingInterface) Foo(ctx context.Context, a int, b int) (err error) {
	err = l.next.Foo(ctx, a, b)

	if err != nil {
		l.log.Log(
			"method", "Foo",
			"error", err.Error(),
		)
	}
	return err
}
```

## Using mongen

Command mongen generates interface implementations that add monitoring. Let's
see an example. Given the following Go interface definition:

```go
package example // import "github.com/Bo0mer/gentools/pkg/example"

type Interface interface {
	Foo(ctx context.Context, a, b int) (err error)
	Bar(ctx context.Context, c string) (x string, y int, err error)
}
```

`mogen` needs the following invocation:

`$ mogen github.com/Bo0mer/gentools/pkg/example Interface | gofmt`

and will produce the following code:

```go
package example

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

// NewMonitoringInterface emits metrics for executed operations. The number of
// total operations is accumulated in totalOps, while the number of failed
// operations is accumulated in failedOps. In addition, the duration for each
// operation (no matter whether it failed or not) is recorded in opsDuration.
// All measurements are labeled by operation name, thus the metrics should have
// a single label field 'operation'.
func NewMonitoringInterface(next Interface, totalOps, failedOps metrics.Counter, opsDuration metrics.Histogram) Interface {
	return &monitoringInterface{
		totalOps:    totalOps,
		failedOps:   failedOps,
		opsDuration: opsDuration,
		next:        next,
	}
}

// Generated using github.com/Bo0mer/gentools/cmd/mongen.
type monitoringInterface struct {
	totalOps    metrics.Counter
	failedOps   metrics.Counter
	opsDuration metrics.Histogram
	next        Interface
}

func (m *monitoringInterface) Bar(ctx context.Context, c string) (x string, y int, err error) {
	start := time.Now()
	x, y, err = m.next.Bar(ctx, c)

	m.totalOps.With("operation", "bar").Add(1)
	m.opsDuration.With("operation", "bar").Observe(time.Since(start).Seconds())

	if err != nil {
		m.failedOps.With("operation", "bar").Add(1)
	}
	return x, y, err
}

func (m *monitoringInterface) Foo(ctx context.Context, a int, b int) (err error) {
	start := time.Now()
	err = m.next.Foo(ctx, a, b)

	m.totalOps.With("operation", "foo").Add(1)
	m.opsDuration.With("operation", "foo").Observe(time.Since(start).Seconds())

	if err != nil {
		m.failedOps.With("operation", "foo").Add(1)
	}
	return err
}
```
