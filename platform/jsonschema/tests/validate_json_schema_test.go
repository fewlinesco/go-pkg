package tests

import (
	_ "embed"
	"errors"
	"github.com/fewlinesco/go-pkg/platform/jsonschema"

	"reflect"
	"testing"
)

type expectedModel struct {
	ID       string            `json:"id"`
	Code     string            `json:"code"`
	DataType string            `json:"datatype"`
	Name     map[string]string `json:"name"`
}

type validateAgainstJsonSchemaTestData struct {
	Name             string
	JSONInput        string
	ValidateResponse func(t *testing.T, outcome error, data validateAgainstJsonSchemaTestData)
	ExpectedError    error
}

var testCases = []validateAgainstJsonSchemaTestData{
	{
		Name:          "when_the_decoding_happens_without_an_error",
		JSONInput:     `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer"}`,
		ExpectedError: nil,
	},
	{
		Name:      "when_a_parameter_has_an_incorrect_datatype",
		JSONInput: `{"code": 1, "id": "815b73a1-3d89-4c68-a4d8-1f36c091a533", "datatype": "string"}`,
		ExpectedError: jsonschema.ValidationError{
			Details: map[string]string{
				"code": "Invalid type. Expected: string, given: integer",
			},
		},
	},
	{
		Name:      "when_a_parameter_has_an_incorrect_enum_type",
		JSONInput: `{"code": "code", "id": "10fbd107-4bcf-4c91-8ee2-957e07d6109e", "datatype": "hello"}`,
		ExpectedError: jsonschema.ValidationError{
			Details: map[string]string{
				"datatype": `datatype must be one of the following: "string", "boolean", "localizedString", "integer", "number"`,
			},
		},
	},
	{
		Name:      "when_the_json_has_an_unknown_field_and_the_decoder_options_are_empty",
		JSONInput: `{"code": "code", "id": "78c8803e-ce4e-474e-97c4-7bd6d565ddca", "datatype": "string", "unknown_field": "hello"}`,
		ExpectedError: jsonschema.AdditionalPropertyError{
			Details: map[string]string{
				"unknown_field": "Additional property unknown_field is not allowed",
			},
		},
	},
	{
		Name:      "when_the_json_has_an_unknown_field_and_the_decoder_options_specify_it_should_not_allow_unknown_fields",
		JSONInput: `{"code": "code", "id": "ec85bd34-67bf-4418-95cb-2616e914bfc9", "datatype": "string", "unknown_field": "hello"}`,
		ExpectedError: jsonschema.AdditionalPropertyError{
			Details: map[string]string{
				"unknown_field": "Additional property unknown_field is not allowed",
			},
		},
	},
	{
		Name:      "it_returns_a_bad_request_reponse_when_a_required_key_is_missing",
		JSONInput: `{"code": "code", "datatype": "integer"}`,
		ExpectedError: jsonschema.MissingPropertyError{
			Details: map[string]string{
				"id": "id is required",
			},
		},
	},
	{
		Name:      "it_returns_a_bad_request_reponse_when_a_required_key_is_missing_and_there_is_an_issue_with_another_property",
		JSONInput: `{"code": 5, "datatype": "integer"}`,
		ExpectedError: jsonschema.MissingPropertyError{
			Details: map[string]string{
				"id":   "id is required",
				"code": "Invalid type. Expected: string, given: integer",
			},
		},
	},
	{
		Name:          "it_can_properly_validate_the_localized_string",
		JSONInput:     `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"en-US": "this is a test", "fr-FR": "another test"}}`,
		ExpectedError: nil,
	},
	{
		Name:      "it_throws_an_error_when_a_required_nested_property_is_missing",
		JSONInput: `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"fr-FR": "another test"}}`,
		ExpectedError: jsonschema.MissingPropertyError{
			Details: map[string]string{
				"name": "en-US is required",
			},
		},
	},
	{
		Name:      "it_throws_an_error_when_a_property_with_invalid_format_is_added",
		JSONInput: `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer", "name": {"en-US": "this is a test", "French": "ceci est une test"}}`,
		ExpectedError: jsonschema.AdditionalPropertyError{
			Details: map[string]string{
				"name": "Additional property French is not allowed",
			},
		},
	},
}

//go:embed testdata/json_schema_with_definition.json
var jsonSchema []byte

func TestValidateJSONAgainstSchema(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			if err := jsonschema.ValidateJSONAgainstSchema([]byte(tc.JSONInput), jsonSchema); err != nil {
				if tc.ExpectedError == nil {
					t.Fatalf("the request body failed the validation but the test did not expect this: %v", err)
				}
				unwrappedError := errors.Unwrap(err)
				if reflect.TypeOf(tc.ExpectedError).Name() != reflect.TypeOf(unwrappedError).Name() {
					t.Fatalf("the error return from the JSON schema validator is not of the expected type. Expected: %T, got %T", tc.ExpectedError, unwrappedError)
				}

				returnedErrorDetails := getErrorDetails(unwrappedError)
				expectedErrorDetails := getErrorDetails(tc.ExpectedError)

				if len(expectedErrorDetails) != len(returnedErrorDetails) {
					t.Fatalf("the returned error's number of details does not match the expectation.\n\tExpected: %+v\n\tReturned: %+v", expectedErrorDetails, returnedErrorDetails)
				}

				for key, message := range returnedErrorDetails {
					if expectedMessage, ok := expectedErrorDetails[key]; !ok || expectedMessage != message {
						t.Fatalf("the error message  detail for the key: '%s' does not match the expectation.\n\tExpected message: %s\n\tReturned message: %s", key, expectedMessage, message)
					}
				}
				return
			}
		})
	}
}

func getErrorDetails(err error) map[string]string {
	var errorDetails map[string]string
	if validationErr, ok := err.(jsonschema.ValidationError); ok {
		errorDetails = validationErr.Details
	}
	if additionalPropertyErr, ok := err.(jsonschema.AdditionalPropertyError); ok {
		errorDetails = additionalPropertyErr.Details
	}
	if missingPropertyErr, ok := err.(jsonschema.MissingPropertyError); ok {
		errorDetails = missingPropertyErr.Details
	}
	return errorDetails
}
