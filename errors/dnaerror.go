package errors

type STCError struct {
	errmsg string
	callstack *CallStack
	root error
	code ErrCode
}

func (e STCError) Error() string {
	return e.errmsg
}

func (e STCError) GetErrCode()  ErrCode {
	return e.code
}

func (e STCError) GetRoot()  error {
	return e.root
}

func (e STCError) GetCallStack()  *CallStack {
	return e.callstack
}
