package errors

import (
	stderrs "errors"
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
)

type stackType struct {
	programCounters []uintptr
	frames          *runtime.Frames
}

func funcName(fullyQualifiedFunctionName string) string {
	return path.Base(fullyQualifiedFunctionName)
}

func (s *stackType) formatVerboseStackTrace(w io.Writer) {
	// Lazy load the frames when someone actually wants the stack trace.
	if s.frames == nil {
		s.frames = runtime.CallersFrames(s.programCounters)
	}
	for {
		frame, more := s.frames.Next()
		io.WriteString(w, funcName(frame.Function))
		io.WriteString(w, "\n\t")
		io.WriteString(w, frame.File)
		io.WriteString(w, ":")
		io.WriteString(w, strconv.Itoa(frame.Line))
		io.WriteString(w, "\n")
		if !more {
			break
		}
	}
}

// wrappedErrorWithStackTrace is for wrapping an error. We wrap the error and
// grab the stack trace for it.
type wrappedErrorWithStackTrace struct {
	err error
	stk *stackType
}

func (e *wrappedErrorWithStackTrace) Unwrap() error {
	return Unwrap(e.err)
}

// Error just returns the Error().
func (e *wrappedErrorWithStackTrace) Error() string {
	return e.err.Error()
}

func (e *wrappedErrorWithStackTrace) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			io.WriteString(state, e.err.Error())
			// TODO: Unravel all the errors.
			io.WriteString(state, "\n")
			// TODO: Unravel all the stack traces.
			e.stk.formatVerboseStackTrace(state)
			return
		}
		fallthrough
	case 's':
		io.WriteString(state, e.err.Error())
	case 'q':
		fmt.Fprintf(state, "%q", e.err.Error())
	}
}

// errorWithStackTrace is the base case for an error. This is when we create a
// new error, if you are wrapping an error then use wrappedErrorWithStackTrace.
type errorWithStackTrace struct {
	msg string
	stk *stackType
}

// Error just returns the Error().
func (e *errorWithStackTrace) Error() string {
	return e.msg
}

func (e *errorWithStackTrace) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			io.WriteString(state, e.msg)
			io.WriteString(state, "\n")
			e.stk.formatVerboseStackTrace(state)
			return
		}
		fallthrough
	case 's':
		io.WriteString(state, e.msg)
	case 'q':
		fmt.Fprintf(state, "%q", e.msg)
	}
}

// getProgramCountersForPackage returns the program counters in context of this
// package (skipping the callers internal to this package)
func getProgramCountersForPackage() []uintptr {
	// We need to choose the maximum depth in the stack.
	//
	// Most Go programs don't go deep in to the call stack. Most of the time
	// engineers are looking at the first line; in other cases they may look for
	// some context. Ideally, engineers are using logs and error messages to
	// provide the context; but sometimes rely on the stack trace.
	//
	// In those case, we need to ensure that a sufficient stack trace is
	// available. In my experience, it is rare to need more than a handful of
	// calls before you are able to determine the call path that you need to
	// focus on debugging. So, we will limit this to a smaller number.
	//
	// 32 was chosen because of the above and to match github.com/pkg/errors
	const maxStackTraceDepth = 32
	var programCounter [maxStackTraceDepth]uintptr

	// We use 3 to skip the right amount of callers; we want to skip all the
	// internal function calls to this package. That way the users first line is
	// where they created the error.
	//
	//  1. runtime.Callers
	//  2. getProgramCountersForPackage
	//  3. NewWithStackTrace
	const skipInternalFunctions = 3
	actualDepth := runtime.Callers(skipInternalFunctions, programCounter[:])

	// We want to set the capacity to avoid ownership/mutability issues.
	return programCounter[0:actualDepth:actualDepth]
}

// Wrap is a proxy to fmt.Errorf — I know this is in a different package but
// it is _very_ annoying and the API is clearly for errors.
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

func Wrapf(err error, format string, a ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), err)
}

func WrapWithStackTrace(err error, message string) error {
	return &wrappedErrorWithStackTrace{
		err: Wrap(err, message),
		stk: &stackType{
			programCounters: getProgramCountersForPackage(),
			frames:          nil,
		},
	}
}

func WrapWithStackTracef(err error, format string, a ...any) error {
	return &wrappedErrorWithStackTrace{
		err: Wrapf(err, format, a...),
		stk: &stackType{
			programCounters: getProgramCountersForPackage(),
			frames:          nil,
		},
	}
}

// As is a proxy to standard errors.As
func As(err error, target any) bool {
	return stderrs.As(err, target)
}

// Is is a proxy to standard errors.Is
func Is(err error, target error) bool {
	return stderrs.Is(err, target)
}

// Join is a proxy to standard errors.Join
func Join(err ...error) error {
	return stderrs.Join(err...)
}

// New is a proxy to standard errors.New
func New(text string) error {
	return stderrs.New(text)
}

// Newf is a proxy to fmt.Errorf
func Newf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

func NewWithStackTrace(text string) error {
	return &errorWithStackTrace{
		msg: text,
		stk: &stackType{
			programCounters: getProgramCountersForPackage(),
			frames:          nil,
		},
	}
}

func NewWithStackTracef(format string, a ...any) error {
	return &errorWithStackTrace{
		msg: fmt.Sprintf(format, a...),
		stk: &stackType{
			programCounters: getProgramCountersForPackage(),
			frames:          nil,
		},
	}
}

// Unwrap is a proxy to errors.Unwrap
func Unwrap(err error) error {
	return stderrs.Unwrap(err)
}
