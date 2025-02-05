# Errors

Errors is a go package that:

1. Wraps the standard error interface from "errors" and "fmt".
2. Brings the stack traces from `github.com/pkg/errors`

## Installation

Use Go to install the dependency.

```shell
go get github.com/scottgreenup/errors
```

## Usage

```go
// Just like https://pkg.go.dev/errors#New
New("file not found")

// Just like https://pkg.go.dev/fmt#Errorf
Newf("file %s was not found", "/etc/passwd")

// A better interface for wrapping
errA := New("bottom")
errB := Wrap(errA, "middle")
errC := Wrapf(errB, "top %q", "some details")

fmt.Println(errC)                               // top "some details": middle: bottom
fmt.Println(errors.Unwrap(errC))                // middle: bottom
fmt.Println(errors.Unwrap(errors.Unwrap(errC))) // bottom

// We also get stack tracing
err := New("original error")
errWithStackTracing := WrapWithStackTrace(err, "some information")
// some information: original error
// errors.TestFun
//      ...errors/errors_test.go:30
// testing.tRunner
//      .../testing/testing.go:1690
// runtime.goexit
//      ...runtime/asm_arm64.s:1223
fmt.Printf("%+v\n", errWithStackTracing)
fmt.Println(Is(errWithStackTracing, err)) // true
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## Why?

Personally, it is the tiny and totally not important thing of having to use two
different packages (i.e. "errors" and "fmt"). I also want stack tracing in some
projects.