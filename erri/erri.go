package erri

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type erriItem struct {
	Trace string
	Info  string
	Err   error
}

type erri struct {
	chain    []erriItem
	info     string
	infoIi8n string
}

func (e erri) Error() string {
	var builder strings.Builder
	builder.WriteString("Catch Error: ")
	builder.WriteString(e.chain[0].Info)
	builder.WriteString(", Trace info: \n")
	for _, item := range e.chain {
		builder.WriteByte('\t')
		builder.WriteString(item.Trace)
		builder.WriteByte('\t')
		if item.Info != "" {
			builder.WriteString("Error Info: ")
			builder.WriteString(item.Info)
		}
		builder.WriteString("\n")
	}
	if e.info != "" {
		builder.WriteString("Extra Error Info: ")
		builder.WriteString(e.info)
		builder.WriteString("\n")
	}

	return builder.String()
}

func callStackTrace() (pos string) {
	pc, _, _, _ := runtime.Caller(2)
	caller := runtime.FuncForPC(pc)
	file, line := caller.FileLine(pc)
	fn := caller.Name()
	pos = fmt.Sprintf("%s:%d\t%s", file, line, fn)
	return
}
func callStackTraceWithLine(line string) (pos string) {
	pc, _, _, _ := runtime.Caller(2)
	caller := runtime.FuncForPC(pc)
	fn := caller.Name()
	pos = fmt.Sprintf("%s\t%s", line, fn)
	return
}

func PrimitiveError(e error) error {
	if err, ok := e.(erri); ok {
		return err.chain[0].Err
	}
	return e
}

// 如果要放入Info消息，使用这个方法，可以用ErrorInfo提取出错误。用于前端接口返回，支持I18n
func WithInfo(e error, info string) error {
	if e == nil {
		return nil
	}
	trace := callStackTrace()
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace})
		err.info = info
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Err: e, Info: e.Error()})
	err.info = info
	return err
}

// 如果要放入Info消息，使用这个方法，可以用ErrorInfo提取出错误。用于前端接口返回，支持I18n
func WithInfoWithLine(line string, e error, info string) error {
	if e == nil {
		return nil
	}
	trace := callStackTraceWithLine(line)
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace})
		err.info = info
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Err: e, Info: e.Error()})
	err.info = info
	return err
}

// 用于前端API，如果有Info，会根据Info，返回Info，如果没有，返回底层Err. 会返回I18N和Info的两条错误信息。
func ErrorInfo(e error) string {
	if e == nil {
		return ""
	}
	var errInfo string
	if err, ok := e.(erri); ok {
		if err.info != "" {
			errInfo = err.info
		} else {
			p := PrimitiveError(e)
			if p == nil {
				return ""
			}
			errInfo = p.Error()
		}
	} else {
		errInfo = e.Error()
	}
	return errInfo
}

// Error returns an error which prefixes function name, similar to the call-stack
func Error(e error) error {
	if e == nil {
		return nil
	}
	trace := callStackTrace()
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace})
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Err: e, Info: e.Error()})
	return err
}

// Error returns an error which prefixes function name, similar to the call-stack
func ErrorWithLine(line string, e error) error {
	if e == nil {
		return nil
	}
	trace := callStackTraceWithLine(line)
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace})
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Err: e, Info: e.Error()})
	return err
}

// ErrorWrap returns an error which combine current function name and the call-stack
func ErrorWrap(e error, format string, a ...interface{}) error {
	trace := callStackTrace()
	if e == nil {
		if format == "" {
			return nil
		}
		info := fmt.Sprintf(format, a...)
		err := erri{}
		err.chain = append(err.chain, erriItem{Trace: trace, Info: info, Err: errors.New(info)})
		return err
	}
	info := fmt.Sprintf(format, a...)
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace, Info: info})
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Info: e.Error(), Err: e})
	err.chain = append(err.chain, erriItem{Trace: trace, Info: info})
	return err
}

// ErrorWrap returns an error which combine current function name and the call-stack
func ErrorWrapWithLine(line string, e error, format string, a ...interface{}) error {
	trace := callStackTraceWithLine(line)
	if e == nil {
		if format == "" {
			return nil
		}
		info := fmt.Sprintf(format, a...)
		err := erri{}
		err.chain = append(err.chain, erriItem{Trace: trace, Info: info, Err: errors.New(info)})
		return err
	}
	info := fmt.Sprintf(format, a...)
	if err, ok := e.(erri); ok {
		err.chain = append(err.chain, erriItem{Trace: trace, Info: info})
		return err
	}
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Info: e.Error(), Err: e})
	err.chain = append(err.chain, erriItem{Trace: trace, Info: info})
	return err
}

// Errorf returns an error which also prefixes the function name and formats the current error
func Errorf(format string, a ...interface{}) error {
	trace := callStackTrace()
	info := fmt.Sprintf(format, a...)
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Info: info, Err: errors.New(info)})
	return err
}

// Errorf returns an error which also prefixes the function name and formats the current error
func ErrorfWithLine(line string, format string, a ...interface{}) error {
	trace := callStackTraceWithLine(line)
	info := fmt.Sprintf(format, a...)
	err := erri{}
	err.chain = append(err.chain, erriItem{Trace: trace, Info: info, Err: errors.New(info)})
	return err
}
