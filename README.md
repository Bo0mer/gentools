# gentools

Gentools provides a collection of tools that help by generating Go interface
implementations that add logging or tracing. 

## Using tracegen

Given the following Go interface definition:

```go
type Interface interface {
	SomethingThatNeedsToBeTraced(ctx context.Context, n int) error
	NoTraceHere(n int) error
}
```

`tracegen` will produce the following generated code:

`$ tracegen github.com/Bo0mer/gentools/pkg/example Interface | gofmt`

```go
type tracingInterface struct {
	next example.Interface
}

func (l *tracingInterface) NoTraceHere(n int) error {
	return l.next.NoTraceHere(n)
}

func (l *tracingInterface) SomethingThatNeedsToBeTraced(ctx context.Context, n int) error {
	ctx, span := trace.StartSpan(ctx, "github.com/Bo0mer/gentools/pkg/example.Interface")
	defer span.End()

	return l.next.SomethingThatNeedsToBeTraced(ctx, n)
}
```

The idea is that you just copy-paste the generated code into your source base.
If any additional context information should be added to the traces, just edit
the generated code before saving it.

## Using logen

Given the following Go interface definition:

```go
type Interface interface {
	SomethingThatNeedsToBeTraced(ctx context.Context, n int) error
	NoTraceHere(n int) error
}
```

`logen` will produce the following generated code:

`$ logen github.com/Bo0mer/gentools/pkg/example Interface | gofmt`

```go
type loggingInterface struct {
	log  logrus.FieldLogger
	next example.Interface
}

func (l *loggingInterface) NoTraceHere(n int) (ret0 error) {
	start := time.Now()
	defer func() {
		l.log.WithFields(logrus.Fields{
			"method": "NoTraceHere",
			"took":   time.Since(start).Seconds(),
			"in_n":   n,
			"error":  ret0,
		})
	}()
	return l.next.NoTraceHere(n)
}

func (l *loggingInterface) SomethingThatNeedsToBeTraced(ctx context.Context, n int) (ret0 error) {
	start := time.Now()
	defer func() {
		l.log.WithFields(logrus.Fields{
			"method": "SomethingThatNeedsToBeTraced",
			"took":   time.Since(start).Seconds(),
			"in_n":   n,
			"error":  ret0,
		})
	}()
	return l.next.SomethingThatNeedsToBeTraced(ctx, n)
}
```
