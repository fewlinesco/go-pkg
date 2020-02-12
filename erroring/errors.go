package erroring

type Source string

type Kind string

type Operation string

func (o Operation) String() string {
	return string(o)
}

const (
	KindUnauthorized             Kind = "unauthorized"
	KindInconsistentIndempotency      = "inconsistent_idempotency"
	KindNotFound                      = "not_found"
	KindUnparsable                    = "unparsable_format"
	KindMissingRequiredArguments      = "missing_required_arguments"
	KindUnprocessablePayload          = "unprocessable_payload"
	KindRemoteFailure                 = "remote_failure"
	KindUnexpected                    = "unexpected"

	SourceClient  Source = "client"
	SourceServer         = "server"
	SourceNetwork        = "network"
	SourceMe             = "application"
	SourceUnknown        = "unknown"
)

type Error struct {
	Kind         Kind
	Operation    Operation
	Source       Source
	Err          error
	RelevantData map[string]string
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Stacktrace() []string {
	ops := []string{string(e.Operation)}

	err, ok := e.Unwrap().(*Error)
	if !ok {
		return ops
	}

	ops = append(ops, err.Stacktrace()...)

	return ops
}

func (e Error) Unwrap() error {
	if e.Err == nil {
		return nil
	}

	_, ok := e.Err.(*Error)
	if !ok {
		return nil
	}

	return e.Err
}

type BusinessError interface {
	error
	BusinessError() string
	Summary() string
	Detail() map[string]string
}

type InternalError interface {
	error
	InternalError() string
}

type Business struct {
	Message string
	Detail  map[string]string
}

func (b Business) Error() string {
	return b.Message
}

func (b Business) BusinessError() string {
	return b.Error()
}

type Internal struct {
	Message string
}

func NewInternal(msg string) Internal {
	return Internal{Message: msg}
}

func (i Internal) Error() string {
	return i.Message
}

func (b Internal) InternalError() string {
	return b.Error()
}
