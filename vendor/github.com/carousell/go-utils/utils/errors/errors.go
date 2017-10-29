package errors

import (
	errs "errors"
	"runtime"
	"strings"
)

var (
	basePath = ""
)

type StackFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"function"`
}

type ErrorExt interface {
	error
	Callers() []uintptr
	StackFrame() []StackFrame
	Cause() error
}

type customError struct {
	Msg   string
	stack []uintptr
	frame []StackFrame
	cause error
}

func (c customError) Error() string {
	return c.Msg
}

func (c customError) Callers() []uintptr {
	return c.stack[:]
}

func (c customError) StackFrame() []StackFrame {
	return c.frame
}

func (c customError) Cause() error {
	return c.cause
}

func (c *customError) generateStack(skip int) []StackFrame {
	stack := []StackFrame{}
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
	}
	c.frame = stack
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

func New(msg string) error {
	return errs.New(msg)
}

func NewWithSkip(msg string, skip int) ErrorExt {
	return WrapWithSkip(errs.New(msg), "", skip+1)
}

func Wrap(err error, msg string) ErrorExt {
	return WrapWithSkip(err, msg, 1)
}

func WrapWithSkip(err error, msg string, skip int) ErrorExt {
	if err == nil {
		return nil
	}
	msg = strings.TrimSpace(msg)
	if msg != "" {
		msg = msg + " :"
	}
	if e, ok := err.(ErrorExt); ok {
		c := &customError{
			Msg:   msg + e.Error(),
			cause: e.Cause(),
		}
		c.stack = e.Callers()
		c.frame = e.StackFrame()
		return c
	} else {
		c := &customError{
			Msg:   msg + err.Error(),
			cause: err,
		}
		c.generateStack(skip + 1)
		return c
	}
}

func SetBaseFilePath(path string) {
	path = strings.TrimSpace(path)
	if path != "" {
		basePath = path
	}
}
