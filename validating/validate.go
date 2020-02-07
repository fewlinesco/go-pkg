package validating

import (
	"errors"
	"github.com/fewlinesco/go-pkg/erroring"
	"github.com/go-ozzo/ozzo-validation/v4"
	"reflect"
)

func ValidateRequire(operation erroring.Operation, inputPtr interface{}, fields ...interface{}) error {
	rules := make([]*validation.FieldRules, len(fields))
	for index, field := range fields {
		rules[index] = validation.Field(field, validation.Required)
	}

	return toErr(operation, validation.ValidateStruct(inputPtr, rules...),
		erroring.KindMissingRequiredArguments, "missing required parameter")
}

func ValidateBusiness(operation erroring.Operation, inputPtr interface{}, rules ...*validation.FieldRules) error {
	return toErr(operation, validation.ValidateStruct(inputPtr, rules...),
		erroring.KindUnprocessablePayload, "invalid parameters")
}

func ErrAlreadyTaken(operation erroring.Operation, inputPtr interface{}, field interface{}) error {
	rule := validation.Field(field, validation.By(isAlreadyTaken))

	return ValidateBusiness(operation, inputPtr, rule)
}

func isAlreadyTaken(interface{}) error {
	return errors.New("already taken")
}

func toErr(operation erroring.Operation, err error, kind erroring.Kind, reason string) error {
	if err == nil {
		return nil
	}

	e, ok := err.(validation.Errors)
	if !ok {
		return &erroring.Error{
			Operation: operation,
			Kind:      erroring.KindUnexpected,
			Source:    erroring.SourceMe,
			Err:       e,
		}
	}

	return &erroring.Error{
		Operation:    operation,
		Kind:         kind,
		Source:       erroring.SourceClient,
		Err:          errors.New(reason),
		RelevantData: toErrMap(map[string]error(e)),
	}
}

func toErrMap(errors map[string]error) map[string]string {
	errs := make(map[string]string, len(errors))

	for k, v := range errors {
		errs[k] = v.Error()
	}

	return errs
}

type Business struct {
	Ptr    ValidationError
	Fields []*BusinessField
}

type BusinessField struct {
	Ptr   interface{}
	Rules []validation.Rule
}

type ValidationError interface {
	IsValid() bool
	SetIsValid(bool)
}

func BusinessValidations(ptr ValidationError, fields ...*BusinessField) *Business {
	return &Business{Ptr: ptr, Fields: fields}
}

func BusinessValidation(ptr interface{}, rules ...validation.Rule) *BusinessField {
	return &BusinessField{Ptr: ptr, Rules: rules}
}

func ValidateRequired(inputPtr interface{}, requiredPtr ValidationError, businessPtr *Business) {
	inputStruct := reflect.ValueOf(inputPtr).Elem()
	requiredStruct := reflect.ValueOf(requiredPtr).Elem()

	requiredPtr.SetIsValid(true)
	for i := 0; i < requiredStruct.NumField(); i++ {
		requiredField := requiredStruct.Field(i)
		if !requiredField.CanSet() {
			continue
		}

		requireFieldType := requiredStruct.Type().Field(i)
		inputField := inputStruct.FieldByName(requireFieldType.Name)

		if err := validation.Required.Validate(inputField.Interface()); err != nil {
			requiredPtr.SetIsValid(false)
			errValue := err.Error()
			requiredField.Set(reflect.ValueOf(&errValue))
		}
	}

	if !requiredPtr.IsValid() {
		return
	}

	businessStruct := reflect.ValueOf(businessPtr.Ptr).Elem()
	businessPtr.Ptr.SetIsValid(true)
	for _, bField := range businessPtr.Fields {
		bFieldValue := reflect.ValueOf(bField.Ptr)
		bFieldStructField := findStructField(businessStruct, bFieldValue)
		inputField := inputStruct.FieldByName(bFieldStructField.Name)

		if err := validation.Validate(inputField.Interface(), bField.Rules...); err != nil {
			businessPtr.Ptr.SetIsValid(false)
			errValue := err.Error()
			businessStruct.FieldByName(bFieldStructField.Name).Set(reflect.ValueOf(&errValue))
		}
	}
}

func findStructField(structValue reflect.Value, fieldValue reflect.Value) *reflect.StructField {
	ptr := fieldValue.Pointer()
	for i := structValue.NumField() - 1; i >= 0; i-- {
		sf := structValue.Type().Field(i)
		if ptr == structValue.Field(i).UnsafeAddr() {
			// do additional type comparison because it's possible that the address of
			// an embedded struct is the same as the first field of the embedded struct
			if sf.Type == fieldValue.Elem().Type() {
				return &sf
			}
		}
		if sf.Anonymous {
			// delve into anonymous struct to look for the field
			fi := structValue.Field(i)
			if sf.Type.Kind() == reflect.Ptr {
				fi = fi.Elem()
			}
			if fi.Kind() == reflect.Struct {
				if f := findStructField(fi, fieldValue); f != nil {
					return f
				}
			}
		}
	}

	return nil
}
