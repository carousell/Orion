package errors

import (
	"fmt"
	"runtime"
	"strings"

	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
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
	//GRPCStatus allows ErrorExt to be treated as a GRPC Error
	GRPCStatus() *grpcstatus.Status
}

//NotifyExt is the interface definition for notifier related options
type NotifyExt interface {
	ShouldNotify() bool
	Notified(status bool)
}

type customError struct {
	Msg          string
	stack        []uintptr
	frame        []StackFrame
	cause        error
	shouldNotify bool
	status       *grpcstatus.Status
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

func (c customError) GRPCStatus() *grpcstatus.Status {
	if c.status == nil {
		return grpcstatus.New(codes.Internal, c.Error())
	}
	return c.status
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

//New creates a new error with stack information
func New(msg string) ErrorExt {
	return NewWithSkip(msg, 1)
}

//NewWithStatus creates a new error with statck information and GRPC status
func NewWithStatus(msg string, status *grpcstatus.Status) ErrorExt {
	return NewWithSkipAndStatus(msg, 1, status)
}

//NewWithSkip creates a new error skipping the number of function on the stack
func NewWithSkip(msg string, skip int) ErrorExt {
	return WrapWithSkip(fmt.Errorf(msg), "", skip+1)
}

//NewWithSkipAndStatus creates a new error skipping the number of function on the stack and GRPC status
func NewWithSkipAndStatus(msg string, skip int, status *grpcstatus.Status) ErrorExt {
	return WrapWithSkipAndStatus(fmt.Errorf(msg), "", skip+1, status)
}

//Wrap wraps an existing error and appends stack information if it does not exists
func Wrap(err error, msg string) ErrorExt {
	return WrapWithSkip(err, msg, 1)
}

//Wrap wraps an existing error and appends stack information if it does not exists along with GRPC status
func WrapWithStatus(err error, msg string, status *grpcstatus.Status) ErrorExt {
	return WrapWithSkipAndStatus(err, msg, 1, status)
}

//WrapWithSkip wraps an existing error and appends stack information if it does not exists skipping the number of function on the stack
func WrapWithSkip(err error, msg string, skip int) ErrorExt {
	return WrapWithSkipAndStatus(err, msg, skip+1, nil)
}

//WrapWithSkip wraps an existing error and appends stack information if it does not exists skipping the number of function on the stack along with GRPC status
func WrapWithSkipAndStatus(err error, msg string, skip int, status *grpcstatus.Status) ErrorExt {
	if err == nil {
		return nil
	}

	msg = strings.TrimSpace(msg)
	if msg != "" {
		msg = msg + " :"
	}

	if status == nil {
		// try to get status from existing one from error
		if s, ok := grpcstatus.FromError(err); ok {
			status = s
		}
	}

	//if we have stack information reuse that
	if e, ok := err.(ErrorExt); ok {
		c := &customError{
			Msg:    msg + e.Error(),
			cause:  e.Cause(),
			status: status,
		}

		c.stack = e.Callers()
		c.frame = e.StackFrame()
		if n, ok := e.(NotifyExt); ok {
			c.shouldNotify = n.ShouldNotify()
		}
		return c
	}

	c := &customError{
		Msg:          msg + err.Error(),
		cause:        err,
		shouldNotify: true,
		status:       status,
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
