package erroring

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
