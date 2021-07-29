package web

import (
	"fmt"
	"github.com/fewlinesco/gojsonschema"
)

// JSONSchemaAdditionalPropertyError is returned when the validation failed due to additional properties being present in the JSON input.
type JSONSchemaAdditionalPropertyError struct {
	Details ErrorDetails
}

func (err JSONSchemaAdditionalPropertyError) Error() string {
	return "the json input contains unknown keys"
}

// JSONSchemaMissingPropertyError is returned when the validation failed due to missing required properties in the JSON input.
type JSONSchemaMissingPropertyError struct {
	Details ErrorDetails
}

func (err JSONSchemaMissingPropertyError) Error() string {
	return "the json input is missing required keys"
}

// JSONSchemaValidationError is returned when the validation failed
type JSONSchemaValidationError struct {
	Details ErrorDetails
}

func (err JSONSchemaValidationError) Error() string {
	return "the json input is not valid against the json schema"
}

// ValidateJSONAgainstSchema validates a JSON object against a JSON schema and returns validation errors if there are any.
func ValidateJSONAgainstSchema(jsonData []byte, jsonSchema []byte) error {
	return validateJSONAgainstSchema(jsonData, gojsonschema.NewBytesLoader(jsonSchema))
}

func validateJSONAgainstSchema(jsonData []byte, jsonSchema gojsonschema.JSONLoader) error {
	dataLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := gojsonschema.Validate(jsonSchema, dataLoader)
	if err != nil {
		return fmt.Errorf("could not validate the data against the json schema: %v", err)
	}

	if !result.Valid() {
		additionalPropertyErrors := make(ErrorDetails)
		requiredPropertyErrors := make(ErrorDetails)
		otherErrors := make(ErrorDetails)

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
				otherErrors[propertyName] = desc.Description()
			}
		}

		if len(additionalPropertyErrors) > 0 {
			errDetails := additionalPropertyErrors
			for property, errMessage := range otherErrors {
				errDetails[property] = errMessage
			}

			return fmt.Errorf("the json contains unknown keys: %w", JSONSchemaAdditionalPropertyError{Details: errDetails})
		}

		if len(requiredPropertyErrors) > 0 {
			errDetails := requiredPropertyErrors
			for property, errMessage := range otherErrors {
				errDetails[property] = errMessage
			}

			return fmt.Errorf("the json is missing required keys: %w", JSONSchemaMissingPropertyError{Details: errDetails})
		}

		return fmt.Errorf("%w", JSONSchemaValidationError{Details: otherErrors})
	}
	return nil
}
