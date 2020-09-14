package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fewlinesco/go-pkg/platform/web"
)

type decodeWithJSONSchemaTestData struct {
	Name             string
	Body             string
	ValidateResponse func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData)
	JSONSchemaPath   string
}

func TestDecodeWithJSONSchema(t *testing.T) {
	type expectedModel struct {
		ID       string `json:"id"`
		Code     int    `json:"code"`
		DataType string `json:"datatype"`
	}

	tcs := []decodeWithJSONSchemaTestData{
		{
			Name:           "when the decoding happens without an error",
			Body:           `{"code": 1, "id": "1", "datatype": "string"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			ValidateResponse: func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData) {
				if outcome != nil {
					t.Fatalf("did not expect the function to return an error, but got: %+v", outcome)
				}
			},
		},
		{
			Name:           "when the path to the json schema is incorrect",
			Body:           `{"code": 1, "id": "1", "datatype": "string"}`,
			JSONSchemaPath: "../../testdata/json-schema/json_schema_with_definition.json",
			ValidateResponse: func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData) {
				if outcome == nil {
					t.Fatalf("expected an error but got nil")
				}
			},
		},
		{
			Name:           "when a parameter has an incorrect datatype",
			Body:           `{"code": 1, "id": 1, "datatype": "string"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			ValidateResponse: func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData) {
				if outcome == nil {
					t.Fatalf("expected an error but got nil")
				}
			},
		},
		{
			Name:           "when a parameter has an incorrect enum type",
			Body:           `{"code": 1, "id": "1", "datatype": "hello"}`,
			JSONSchemaPath: "../../../testdata/json-schema/json_schema_with_definition.json",
			ValidateResponse: func(testProvider *testing.T, outcome error, data decodeWithJSONSchemaTestData) {
				if outcome == nil {
					t.Fatalf("expected an error but got nil")
				}
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {

			request, err := http.NewRequest(http.MethodPut, "http://localhost:30000", strings.NewReader(tc.Body))
			if err != nil {
				t.Fatalf("Unable to generate request: %+v", err)
			}

			request.Header.Set("Content-Type", "application/json")

			tc.ValidateResponse(t, web.DecodeWithJSONSchema(request, &expectedModel{}, tc.JSONSchemaPath), tc)
		})
	}
}
