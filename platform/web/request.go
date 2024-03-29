package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/xeipuuv/gojsonschema"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

var validate = validator.New()
var translator *ut.UniversalTranslator
var fieldRegex = regexp.MustCompile(`json: unknown field "([^"]+)"`)

// ErrRequestBodyTooLargeMessage is the error returned by the http.MaxBytesReader() when reading from a reader over the set limit
var ErrRequestBodyTooLargeMessage = "http: request body too large"

// DecoderOptions describes a set of options you can pass to alter the behaviour of the decoders
// AllowUnknownFields ensures that the decoder can pass the json even if it contains a field unknown to the strict
type DecoderOptions struct {
	AllowUnknownFields bool
}

func init() {
	enLocale := en.New()
	translator = ut.New(enLocale, enLocale)
	lang, _ := translator.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, lang)
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Decode reads the body of an HTTP request as JSON and fill a struct with its content. It's also in charge of validating the content of the struct based on gopkg.in/go-playground/validator.v9 validation tags.
func Decode(r *http.Request, val interface{}) error {
	return decode(r, val, DecoderOptions{AllowUnknownFields: false})
}

// DecodeWithJSONSchema takes the path to a json schema and a http request
// And returns an error when the request's payload does not match the JSON schema
func DecodeWithJSONSchema(request *http.Request, model interface{}, filePath string, options DecoderOptions) error {
	_, rootFile, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("%w", NewErrInvalidJSONSchemaFilePath())
	}

	filePath = path.Join(path.Dir(rootFile), filePath)

	jsonSchema := gojsonschema.NewReferenceLoader("file://" + filePath)

	if err := validateRequestPayload(request, model, options, jsonSchema); err != nil {
		return err
	}
	return nil
}

// DecodeWithEmbeddedJSONSchema takes json schema and a http request
// And returns an error when the request's payload does not match the JSON schema
func DecodeWithEmbeddedJSONSchema(request *http.Request, model interface{}, jsonSchemaBytes []byte, options DecoderOptions) error {
	jsonSchema := gojsonschema.NewBytesLoader(jsonSchemaBytes)

	if err := validateRequestPayload(request, model, options, jsonSchema); err != nil {
		return err
	}
	return nil
}

func validateRequestPayload(request *http.Request, model interface{}, options DecoderOptions, jsonSchema gojsonschema.JSONLoader) error {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		if err.Error() == ErrRequestBodyTooLargeMessage {
			return fmt.Errorf("%w", NewErrRequestBodyTooLarge())
		}
		return err
	}
	request.Body = io.NopCloser(bytes.NewBuffer(body))
	payload := gojsonschema.NewBytesLoader(body)

	result, err := gojsonschema.Validate(jsonSchema, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", NewErrBadRequestResponse(nil), err)
	}

	if !result.Valid() {
		additionalPropertyErrors := make(ErrorDetails)
		requiredPropertyErrors := make(ErrorDetails)
		requestContentErrorDetails := make(ErrorDetails)

		for _, desc := range result.Errors() {
			propertyName := desc.Field()
			if propertyName == "(root)" {
				details, ok := desc.Details()["property"]
				if ok {
					propertyName = fmt.Sprintf("%v", details)
				}
			}

			switch desc.Type() {
			case "additional_property_not_allowed":
				additionalPropertyErrors[propertyName] = desc.Description()
			case "required":
				requiredPropertyErrors[propertyName] = desc.Description()
			default:
				requestContentErrorDetails[propertyName] = desc.Description()
			}
		}

		if len(additionalPropertyErrors) > 0 {
			errDetails := additionalPropertyErrors
			for property, errMessage := range requestContentErrorDetails {
				errDetails[property] = errMessage
			}

			return fmt.Errorf("the request body contains unknown keys: %w", NewErrBadRequestResponse(errDetails))
		}

		if len(requiredPropertyErrors) > 0 {
			errDetails := requiredPropertyErrors
			for property, errMessage := range requestContentErrorDetails {
				errDetails[property] = errMessage
			}

			return fmt.Errorf("the request body contains unknown keys: %w", NewErrBadRequestResponse(errDetails))
		}

		return fmt.Errorf("%w", NewErrInvalidRequestBodyContent(requestContentErrorDetails))
	}

	if err := decode(request, model, options); err != nil {
		return err
	}

	return nil
}

func decode(r *http.Request, val interface{}, options DecoderOptions) error {
	decoder := json.NewDecoder(r.Body)

	if !options.AllowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(val); err != nil {
		if err.Error() == ErrRequestBodyTooLargeMessage {
			return fmt.Errorf("%w", NewErrRequestBodyTooLarge())
		}
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			return fmt.Errorf("%v: %w", err, NewErrBadRequestResponse(ErrorDetails{
				e.Field: fmt.Sprintf("%s must be a %s", e.Field, e.Type.String()),
			}))
		case *json.SyntaxError:
			return fmt.Errorf("%v: %w", err, NewErrUnmarshallableJSON())
		}

		if err.Error() == "EOF" {
			return fmt.Errorf("%v: %w", err, NewErrMissingRequestBody())
		}

		if strings.Contains(err.Error(), "json: unknown field") {
			matches := fieldRegex.FindStringSubmatch(err.Error())
			fieldName := matches[1]

			return fmt.Errorf("%v: %w", err, NewErrBadRequestResponse(ErrorDetails{
				fieldName: fmt.Sprintf("%s field is not allowed", fieldName),
			}))
		}

		return fmt.Errorf("%T, %v: %w", err, err, NewErrUnmarshallableJSON())
	}

	return Validate(val, NewErrInvalidRequestBodyContent)
}

// Validate checks the struct is valid based on gopkg.in/go-playground/validator.v9 validation tags.
func Validate(val interface{}, errBuilder func(ErrorDetails) error) error {
	if err := validate.Struct(val); err != nil {
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		lang, _ := translator.GetTranslator("en")

		details := make(ErrorDetails)
		for _, verror := range verrors {
			details[verror.Field()] = verror.Translate(lang)
		}

		return fmt.Errorf("%w", errBuilder(details))
	}

	return nil
}
