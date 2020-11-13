package tests

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/fewlinesco/go-pkg/platform/web"
)

type decodeWithJSONSchemaTestData struct {
	Name             string
	Body             string
	ValidateResponse func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData)
	JSONSchemaPath   string
	DecoderOptions web.DecoderOptions
	ExpectedError bool
	ExpectedOutcome expectedModel
}

type expectedModel struct {
	ID       string `json:"id"`
	Code     string    `json:"code"`
	DataType string `json:"datatype"`
}

func TestDecodeWithJSONSchema(t *testing.T) {

	tcs := []decodeWithJSONSchemaTestData{
		{
			Name:           "when the decoding happens without an error",
			Body:           `{"code": "code", "id": "d43c45b0-f420-4de9-8745-6e3840ab39fd", "datatype": "integer"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{},
			ExpectedError: false,
			ExpectedOutcome: expectedModel{
				ID:       "d43c45b0-f420-4de9-8745-6e3840ab39fd",
				Code:     "code",
				DataType: "integer",
			},
		},
		{
			Name:           "when the path to the json schema is incorrect",
			Body:           `{"code": "code", "id": "c9ecb26a-20ab-4acb-b34e-444457b06b3b", "datatype": "string"}`,
			JSONSchemaPath: "../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{},
			ExpectedError: true,
		},
		{
			Name:           "when a parameter has an incorrect datatype",
			Body:           `{"code": 1", "id": "815b73a1-3d89-4c68-a4d8-1f36c091a533", "datatype": "string"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{},
			ExpectedError: true,
		},
		{
			Name:           "when a parameter has an incorrect enum type",
			Body:           `{"code": "code", "id": "10fbd107-4bcf-4c91-8ee2-957e07d6109e", "datatype": "hello"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{},
			ExpectedError: true,
		},
		{
			Name:           "when the json has an unknown field and the decoder options are empty",
			Body:           `{"code": "code", "id": "78c8803e-ce4e-474e-97c4-7bd6d565ddca", "datatype": "string", "unknown_field": "hello"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{},
			ExpectedError: true,
		},
		{
			Name:           "when the json has an unknown field and the decoder options specify it should not allow unknown fields ",
			Body:           `{"code": "code", "id": "ec85bd34-67bf-4418-95cb-2616e914bfc9", "datatype": "string", "unknown_field": "hello"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{AllowUnknownFields: false},
			ExpectedError: true,
		},
		{
			Name:           "when the json has an unknown field and the decoder options specify it should not allow unknown fields ",
			Body:           `{"code": "code", "id": "8321308a-4cae-4175-8c56-2db087e5ca10", "datatype": "integer", "unknown_field": "hello"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			DecoderOptions: web.DecoderOptions{AllowUnknownFields: true},
			ExpectedError: false,
			ExpectedOutcome: expectedModel{
				ID:       "8321308a-4cae-4175-8c56-2db087e5ca10",
				Code:     "code",
				DataType: "integer",
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodPut, "http://localhost:30000", strings.NewReader(tc.Body))
			if err != nil {
				t.Fatalf("Unable to generate request: %+v", err)
			}

			request.Header.Set("Content-Type", "application/json")

			var result expectedModel
			if err := web.DecodeWithJSONSchema(request, &result, tc.JSONSchemaPath, tc.DecoderOptions); err != nil {
				if !tc.ExpectedError {
					t.Fatalf("the decoder returned an error message but this was not expected: %v", err)
				}
				return
			}

			if !reflect.DeepEqual(result, tc.ExpectedOutcome) {
				t.Fatalf("the expected outcome and result don't match.\n\tExpected: %+v\n\tReceived: %v", tc.ExpectedOutcome, result)
			}
		})
	}
}
