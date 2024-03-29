package tests

import (
	_ "embed"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/fewlinesco/go-pkg/platform/web"
)

type expectedModel struct {
	ID       string            `json:"id"`
	Code     string            `json:"code"`
	DataType string            `json:"datatype"`
	Name     map[string]string `json:"name"`
}

type decodeWithJSONSchemaTestData struct {
	Name             string
	Body             string
	ValidateResponse func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData)
	DecoderOptions   web.DecoderOptions
	ExpectedError    error
	ExpectedOutcome  expectedModel
}

var testCases = []decodeWithJSONSchemaTestData{
	{
		Name:           "when_the_decoding_happens_without_an_error",
		Body:           `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError:  nil,
		ExpectedOutcome: expectedModel{
			ID:       "d43c45b0-f420-4de9-8745-6e3840ab39fd",
			Code:     "code",
			DataType: "integer",
		},
	},
	{
		Name:           "when_a_parameter_has_an_incorrect_datatype",
		Body:           `{"code": 1, "id": "815b73a1-3d89-4c68-a4d8-1f36c091a533", "datatype": "string"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrInvalidRequestBodyContent(web.ErrorDetails{
			"code": "Invalid type. Expected: string, given: integer",
		}),
	},
	{
		Name:           "when_a_parameter_has_an_incorrect_enum_type",
		Body:           `{"code": "code", "id": "10fbd107-4bcf-4c91-8ee2-957e07d6109e", "datatype": "hello"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrInvalidRequestBodyContent(web.ErrorDetails{
			"datatype": `datatype must be one of the following: "string", "boolean", "localizedString", "integer", "number"`,
		}),
	},
	{
		Name:           "when_the_json_has_an_unknown_field_and_the_decoder_options_are_empty",
		Body:           `{"code": "code", "id": "78c8803e-ce4e-474e-97c4-7bd6d565ddca", "datatype": "string", "unknown_field": "hello"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"unknown_field": "Additional property unknown_field is not allowed",
		}),
	},
	{
		Name:           "when_the_json_has_an_unknown_field_and_the_decoder_options_specify_it_should_not_allow_unknown_fields",
		Body:           `{"code": "code", "id": "ec85bd34-67bf-4418-95cb-2616e914bfc9", "datatype": "string", "unknown_field": "hello"}`,
		DecoderOptions: web.DecoderOptions{AllowUnknownFields: false},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"unknown_field": "Additional property unknown_field is not allowed",
		}),
	},
	{
		Name:           "it_returns_a_bad_request_reponse_when_a_required_key_is_missing",
		Body:           `{"code": "code", "datatype": "integer"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"id": "id is required",
		}),
	},
	{
		Name:           "it_returns_a_bad_request_reponse_when_a_required_key_is_missing_and_there_is_an_issue_with_another_property",
		Body:           `{"code": 5, "datatype": "integer"}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"id":   "id is required",
			"code": "Invalid type. Expected: string, given: integer",
		}),
	},
	{
		Name:           "it_can_properly_validate_the_localized_string",
		Body:           `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"en-US": "this is a test", "fr-FR": "another test"}}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError:  nil,
		ExpectedOutcome: expectedModel{
			ID:       "d43c45b0-f420-4de9-8745-6e3840ab39fd",
			Code:     "code",
			DataType: "integer",
			Name:     map[string]string{"en-US": "this is a test", "fr-FR": "another test"},
		},
	},
	{
		Name:           "it_throws_an_error_when_a_required_nested_property_is_missing",
		Body:           `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"fr-FR": "another test"}}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"name": "en-US is required",
		}),
	},
	{
		Name:           "it_throws_an_error_when_a_property_with_invalid_format_is_added",
		Body:           `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"en-US": "this is a test", "French": "ceci est une test"}}`,
		DecoderOptions: web.DecoderOptions{},
		ExpectedError: web.NewErrBadRequestResponse(web.ErrorDetails{
			"name": "Additional property French is not allowed",
		}),
	},
}

//go:embed testdata/json_schema_with_definition.json
var jsonSchema []byte

func TestDecodeWithEmbeddedJSONSchema(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			request := prepareRequest(t, tc.Body)
			var result expectedModel
			if err := web.DecodeWithEmbeddedJSONSchema(request, &result, jsonSchema, tc.DecoderOptions); err != nil {
				if tc.ExpectedError == nil {
					t.Fatalf("the request body failed the validation but the test did not expect this: %v", err)
				}
				checkError(t, tc.ExpectedError, err)
				return
			}

			if !reflect.DeepEqual(result, tc.ExpectedOutcome) {
				t.Fatalf("the expected outcome and result don't match.\n\tExpected: %+v\n\tReceived: %+v", tc.ExpectedOutcome, result)
			}
		})
	}
}

func TestDecodeWithJSONSchema(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			request := prepareRequest(t, tc.Body)
			var result expectedModel
			if err := web.DecodeWithJSONSchema(request, &result, "testdata/json_schema_with_definition.json", tc.DecoderOptions); err != nil {
				if tc.ExpectedError == nil {
					t.Fatalf("the request body failed the validation but the test did not expect this: %v", err)
				}

				checkError(t, tc.ExpectedError, err)
				return
			}

			if !reflect.DeepEqual(result, tc.ExpectedOutcome) {
				t.Fatalf("the expected outcome and result don't match.\n\tExpected: %+v\n\tReceived: %+v", tc.ExpectedOutcome, result)
			}
		})
	}
}

func TestDecodeWithJSONSchemaWithInvalidFilePath(t *testing.T) {
	t.Run("it returns an error when the file path is incorrect", func(t *testing.T) {
		body := `{"code": "code", "id": "c9ecb26a-20ab-4acb-b34e-444457b06b3b", "datatype": "string"}`
		request := prepareRequest(t, body)
		var result expectedModel
		if err := web.DecodeWithJSONSchema(request, &result, "testdata/json_schema/json_schema_with_definition.json", web.DecoderOptions{}); err != nil {
			checkError(t, web.NewErrBadRequestResponse(nil), err)
			return
		}
		t.Fatalf("the request body passed the validation but it should have returned an error")
	})
}

func prepareRequest(t *testing.T, body string) *http.Request {
	request, err := http.NewRequest(http.MethodPut, "http://localhost:30000", strings.NewReader(body))
	if err != nil {
		t.Fatalf("Unable to generate request: %+v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	return request
}

func checkError(t *testing.T, expectedError error, returnedError error) {
	if expectedError == nil {
		t.Fatalf("the request body failed the validation but the test did not expect this: %v", returnedError)
	}

	returnedWebError, ok := errors.Unwrap(returnedError).(*web.Error)
	if !ok {
		t.Fatalf("the decoder should always return a web error but it did not: %v", returnedError)
	}

	expectedWebError, ok := expectedError.(*web.Error)
	if !ok {
		t.Fatalf("the expected error should be a web error but it was unable to cast to one: %v", expectedError)
	}

	if expectedWebError.Code != returnedWebError.Code {
		t.Fatalf("the returned error's code does not match the expectation.\n\tExpected: %+v\n\tReturned: %+v", expectedWebError.Code, returnedWebError.Code)
	}

	if expectedWebError.Message != returnedWebError.Message {
		t.Fatalf("the returned error's message does not match the expectation.\n\tExpected: %+v\n\tReturned: %+v", expectedWebError.Message, returnedWebError.Message)
	}

	if len(expectedWebError.Details) != len(returnedWebError.Details) {
		t.Fatalf("the returned error's number of details does not match the expectation.\n\tExpected: %+v\n\tReturned: %+v", expectedWebError.Details, returnedWebError.Details)
	}

	for key, message := range returnedWebError.Details {
		if expectedMessage, ok := expectedWebError.Details[key]; !ok || expectedMessage != message {
			t.Fatalf("the error message  detail for the key: '%s' does not match the expectation.\n\tExpected message: %s\n\tReturned message: %s", key, expectedMessage, message)
		}
	}
}
