package errors

import (
	"errors"
	"runtime"
	"strings"
)

type Source struct {
	File     string
	Line     int
	FuncName string
}

type Error struct {
	childErr error
	group    string
	code     string
	msg      string
	source   *Source
}

func New(group, code, msg string) Error {
	return Error{
		group: group,
		code:  code,
		msg:   msg,
	}
}

func (err Error) WithError(inErr error) Error {
	err.childErr = inErr
	return err
}

func (err Error) WithMessage(msg string) Error {
	err.msg = msg
	return err
}

func (err Error) WithSource() Error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	funcInfo := runtime.FuncForPC(pc)
	funcNameSplit := strings.Split(funcInfo.Name(), "/")
	funcName := funcNameSplit[len(funcNameSplit)-1]
	err.source = &Source{
		File:     file,
		Line:     line,
		FuncName: funcName,
	}

	return err
}

func (err Error) Group() string {
	return err.group
}

func (err Error) Code() string {
	return err.code
}

func (err Error) Source() *Source {
	return err.source
}

func (err Error) Error() string {
	if err.msg == "" && err.childErr != nil {
		return err.childErr.Error()
	}

	return err.msg
}

func (err Error) Unwrap() error {
	return err.childErr
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func GetError(err error) (Error, bool) {
	var xErr Error
	ok := errors.As(err, &xErr)

	return xErr, ok
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
