package jsonschema

import (
	"fmt"
	"github.com/fewlinesco/gojsonschema"
)

// AdditionalPropertyError is returned when the validation failed due to additional properties being present in the JSON input.
type AdditionalPropertyError struct {
	Details map[string]string
}

func (err AdditionalPropertyError) Error() string {
	return "the json input contains unknown keys"
}

// MissingPropertyError is returned when the validation failed due to missing required properties in the JSON input.
type MissingPropertyError struct {
	Details map[string]string
}

func (err MissingPropertyError) Error() string {
	return "the json input is missing required keys"
}

// ValidationError is returned when the validation failed
type ValidationError struct {
	Details map[string]string
}

func (err ValidationError) Error() string {
	return "the json input is not valid against the JSON schema"
}

// ValidateJSONAgainstSchema validate some json data against a json schema and return detailed validation errors if any is present.
func ValidateJSONAgainstSchema(jsonData []byte, jsonSchema []byte) error {
	dataLoader := gojsonschema.NewBytesLoader(jsonData)
	schemaLoader := gojsonschema.NewBytesLoader(jsonSchema)
	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return fmt.Errorf("could not validate the data against the json schema: %v", err)
	}

	if !result.Valid() {
		additionalPropertyErrors := make(map[string]string)
		requiredPropertyErrors := make(map[string]string)
		otherErrors := make(map[string]string)

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

			return fmt.Errorf("the json contains unknown keys: %w", AdditionalPropertyError{Details: errDetails})
		}

		if len(requiredPropertyErrors) > 0 {
			errDetails := requiredPropertyErrors
			for property, errMessage := range otherErrors {
				errDetails[property] = errMessage
			}

			return fmt.Errorf("the json is missing required keys: %w", MissingPropertyError{Details: errDetails})
		}

		return fmt.Errorf("%w", ValidationError{Details: otherErrors})
	}
	return nil
}
