package errors

import (
	"fmt"
	"runtime"
	"strings"

	// fully extend GRPC Code ability and prevent import cycle
	. "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	basePath = ""
)

//StackFrame represents the stackframe for tracing exception
type StackFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"function"`
}

//ErrorExt is the interface that defines a error, any ErrorExt implementors can use and override errors and notifier package
type ErrorExt interface {
	error
	Callers() []uintptr
	StackFrame() []StackFrame
	//Cause returns the original error object that caused this error
	Cause() error
}

//NotifyExt is the interface definition for notifier related options
type NotifyExt interface {
	ShouldNotify() bool
	Notified(status bool)
}

type GRPCExt interface {
	// TBD: interface support default code and default http status
	// now we're using internal server error as default
	Code() Code
	ToGRPCStatus() status.Status
}

func (c *customError) Code() Code {
	return c.code
}

func (c *customError) ToGRPCStatus() *status.Status {
	return status.New(c.Code(), c.Msg)
}

type customError struct {
	Msg          string
	stack        []uintptr
	frame        []StackFrame
	cause        error
	shouldNotify bool
	code         Code
}

// implements notifier.NotifyExt
func (c *customError) ShouldNotify() bool {
	return c.shouldNotify
}

// implements notifier.NotifyExt
func (c *customError) Notified(status bool) {
	c.shouldNotify = !status
}

func (c customError) Error() string {
	return c.Msg
}

func (c customError) Callers() []uintptr {
	return c.stack[:]
}

func (c customError) StackTrace() []uintptr {
	return c.Callers()
}

func (c customError) StackFrame() []StackFrame {
	return c.frame
}

func (c customError) Cause() error {
	return c.cause
}

func (c *customError) generateStack(skip int) []StackFrame {
	stack := []StackFrame{}
	trace := []uintptr{}
	for i := skip + 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		_, funcName := packageFuncName(pc)
		if basePath != "" {
			file = strings.Replace(file, basePath, "", 1)
		}
		stack = append(stack, StackFrame{
			File: file,
			Line: line,
			Func: funcName,
		})
		trace = append(trace, pc)
	}
	c.frame = stack
	c.stack = trace
	return stack
}

func packageFuncName(pc uintptr) (string, string) {
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "", ""
	}

	packageName := ""
	funcName := f.Name()

	if ind := strings.LastIndex(funcName, "/"); ind > 0 {
		packageName += funcName[:ind+1]
		funcName = funcName[ind+1:]
	}
	if ind := strings.Index(funcName, "."); ind > 0 {
		packageName += funcName[:ind]
		funcName = funcName[ind+1:]
	}

	return packageName, funcName
}

// NewWithCode create a new error with stack trace and code for grpc
func NewWithCode(msg string, code Code) ErrorExt {
	return WrapWithSkipAndCode(fmt.Errorf(msg), "", 2, code)
}

//New creates a new error with stack trace
func New(msg string) ErrorExt {
	return NewWithSkip(msg, 1)
}

//NewWithSkip creates a new error skipping the number of function on the stack
func NewWithSkip(msg string, skip int) ErrorExt {
	return WrapWithSkip(fmt.Errorf(msg), "", skip+1)
}

//Wrap wraps an existing error and appends stack information if it does not exists
func Wrap(err error, msg string) ErrorExt {
	return WrapWithSkip(err, msg, 1)
}

//WrapWithSkip wraps an existing error and appends stack information if it does not exists skipping the number of function on the stack
func WrapWithSkip(err error, msg string, skip int) ErrorExt {
	return WrapWithSkipAndCode(err, msg, skip, MaxCode)
}

func WrapWithSkipAndCode(err error, msg string, skip int, code Code) ErrorExt {
	if err == nil {
		return nil
	}

	msg = strings.TrimSpace(msg)
	if msg != "" {
		msg = msg + " :"
	}

	newCode := Internal // default
	if code != MaxCode {
		newCode = code
	}

	//if we have stack information reuse that
	if e, ok := err.(ErrorExt); ok {
		c := &customError{
			Msg:   msg + e.Error(),
			cause: e.Cause(),
			code:  newCode,
		}
		c.stack = e.Callers()
		c.frame = e.StackFrame()

		if n, ok := e.(NotifyExt); ok {
			c.shouldNotify = n.ShouldNotify()
		}

		if g, ok := e.(GRPCExt); ok {
			if code == MaxCode { // if no new value, keep original code
				c.code = g.Code()
			} else {
				c.code = newCode
			}

		}

		return c
	}

	c := &customError{
		Msg:          msg + err.Error(),
		cause:        err,
		shouldNotify: true,
		code:         newCode,
	}
	c.generateStack(skip + 1)
	return c
}

//SetBaseFilePath sets the base file path for linking source code with reported stack information
func SetBaseFilePath(path string) {
	path = strings.TrimSpace(path)
	if path != "" {
		basePath = path
	}
}
