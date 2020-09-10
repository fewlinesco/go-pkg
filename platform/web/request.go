package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			return fmt.Errorf("%v: %w", err, NewErrBadRequestResponse(ErrorDetails{
				e.Field: fmt.Sprintf("%s must be a %s", e.Field, e.Type.String()),
			}))
		case *json.SyntaxError:
			return fmt.Errorf("%v: %w", err, newErrUnmarshallableJSON())
		}

		if err.Error() == "EOF" {
			return fmt.Errorf("%v: %w", err, newErrMissingRequestBody())
		}

		if strings.Contains(err.Error(), "json: unknown field") {
			matches := fieldRegex.FindStringSubmatch(err.Error())
			fieldName := matches[1]

			return fmt.Errorf("%v: %w", err, NewErrBadRequestResponse(ErrorDetails{
				fieldName: fmt.Sprintf("%s field is not allowed", fieldName),
			}))
		}

		return fmt.Errorf("%T, %v: %w", err, err, newErrUnmarshallableJSON())
	}

	return Validate(val, NewErrBadRequestResponse)
}

// DecodeWithJSONSchema takes the path to a json schema and a http request
// And returns an error when the request's payload does not match the JSON schema
func DecodeWithJSONSchema(request *http.Request, model interface{}, path string) error {
	body, _ := ioutil.ReadAll(request.Body)

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("%w: %v", newErrInvalidJSONSchemaFilePath(), err)
	}

	jsonSchema := gojsonschema.NewReferenceLoader("file://" + absolutePath)
	payload := gojsonschema.NewBytesLoader(body)

	result, err := gojsonschema.Validate(jsonSchema, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", NewErrBadRequestResponse(nil), err)
	}

	if !result.Valid() {
		errorDetails := make(ErrorDetails)

		for _, desc := range result.Errors() {
			errorDetails[desc.Field()] = desc.Description()
		}

		return fmt.Errorf("%w", newErrInvalidRequest(errorDetails))
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := Decode(request, &model); err != nil {
		return err
	}

	return nil
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
