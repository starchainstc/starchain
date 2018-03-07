package errors

import (
	"errors"
)

const callStackDepth = 10

type DetailError interface {
	error
	ErrCoder
	CallStacker
	GetRoot()  error
}


func  NewErr(errmsg string) error {
	return errors.New(errmsg)
}

func NewDetailErr(err error,errcode ErrCode,errmsg string) DetailError{
	if err == nil {return nil}

	STCerr, ok := err.(STCError)
	if !ok {
		STCerr.root = err
		STCerr.errmsg = err.Error()
		STCerr.callstack = getCallStack(0, callStackDepth)
		STCerr.code = errcode

	}
	if errmsg != "" {
		STCerr.errmsg = errmsg + ": " + STCerr.errmsg
	}


	return STCerr
}

func RootErr(err error) error {
	if err, ok := err.(DetailError); ok {
		return err.GetRoot()
	}
	return err
}



