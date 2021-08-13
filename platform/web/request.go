package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/platform/jsonschema"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	en "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
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

// ParseJSONInput takes in a Request, reads and closes the body and ensures it is valid JSON.
func ParseJSONInput(request *http.Request) ([]byte, error) {
	rawJSONInput, err := ioutil.ReadAll(request.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			return nil, fmt.Errorf("parse json input failed: %w", ErrRequestBodyTooLarge)
		}
		return nil, fmt.Errorf("parse json input failed: unable to read the request body: %v", err)
	}
	if err := request.Body.Close(); err != nil {
		return nil, fmt.Errorf("cannot close request body: %v", err)
	}
	if valid := json.Valid(rawJSONInput); !valid {
		return nil, fmt.Errorf("parse json input failed: %w", NewErrInvalidJSON())
	}

	return rawJSONInput, nil
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

	jsonSchema, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("%w", NewErrInvalidJSONSchemaFilePath())
	}

	if err := DecodeWithEmbeddedJSONSchema(request, model, jsonSchema, options); err != nil {
		return err
	}
	return nil
}

// DecodeWithEmbeddedJSONSchema takes json schema and a http request
// And returns an error when the request's payload does not match the JSON schema
func DecodeWithEmbeddedJSONSchema(request *http.Request, model interface{}, jsonSchema []byte, options DecoderOptions) error {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		if err.Error() == ErrRequestBodyTooLargeMessage {
			return fmt.Errorf("%w", ErrRequestBodyTooLarge)
		}
		return fmt.Errorf("could not read request body: %v", err)
	}
	request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := jsonschema.ValidateJSON(body, jsonSchema); err != nil {
		switch unwrappedErr := errors.Unwrap(err).(type) {
		case jsonschema.AdditionalPropertyError:
			return fmt.Errorf("jsonschema validation error: %w", NewErrBadRequestResponse(unwrappedErr.Details))
		case jsonschema.MissingPropertyError:
			return fmt.Errorf("jsonschema validation error: %w", NewErrBadRequestResponse(unwrappedErr.Details))
		case jsonschema.ValidationError:
			return fmt.Errorf("jsonschema validation error: %w", NewErrInvalidRequestBodyContent(unwrappedErr.Details))
		}
	}

	err = json.Unmarshal(body, &model)
	if err != nil {
		return fmt.Errorf("json is valid against json schema, but cannot be unmarshalled to the provided struct: %v", err)
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
			return fmt.Errorf("%w", ErrRequestBodyTooLarge)
		}
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			return fmt.Errorf("%v: %w", err, NewErrBadRequestResponse(ErrorDetails{
				e.Field: fmt.Sprintf("%s must be a %s", e.Field, e.Type.String()),
			}))
		case *json.SyntaxError:
			return fmt.Errorf("%v: %w", err, NewErrInvalidJSON())
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

		return fmt.Errorf("%T, %v: %w", err, err, NewErrInvalidJSON())
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
